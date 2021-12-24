package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	gcontext "mmr/context"
	"mmr/models"
	"mmr/shared"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type tokenPair struct {
	at *tokenDetails
	rt *tokenDetails
}

type tokenDetails struct {
	token   string
	uuid    string
	expires int64
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	//unmarshal user
	var usr models.User
	if err := json.NewDecoder(r.Body).Decode(&usr); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid request: %v\n", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	//validate user
	if err := shared.Validate.Struct(usr); err != nil {
		fmt.Fprintf(os.Stderr, "User validation failed: %v\n", err.(validator.ValidationErrors))
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	//find the user in db
	conn, err := a.p.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer conn.Release()
	row := conn.QueryRow(context.Background(),
		"SELECT id, name, email, pass FROM users WHERE email = $1", usr.Email)

	//serialize db user or return 401 if they don't exist
	var dbUsr models.User
	if err = row.Scan(&dbUsr.Id, &dbUsr.Name, &dbUsr.Email, &dbUsr.Pass); err == pgx.ErrNoRows {
		http.Error(w, "", http.StatusUnauthorized)
		return
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to SELECT: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	//validate password
	if err = dbUsr.ValidatePass(usr.Pass); err != nil {
		fmt.Fprintf(os.Stderr, "Pass is not correct: %v\n", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	//generate and sign jwt
	tp, err := genTokenPair(dbUsr.Id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't sign token: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	//store token pair in redis
	err = storeTokenPair(a.rdb, dbUsr.Id, tp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't store token: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	//marshal and return token pair
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err = json.NewEncoder(w).Encode(map[string]string{
		"access_token":  tp.at.token,
		"refresh_token": tp.rt.token,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (a *App) register(w http.ResponseWriter, r *http.Request) {
	//unmarshal user
	var usr models.User
	if err := json.NewDecoder(r.Body).Decode(&usr); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid request: %v\n", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	//validate user
	if err := shared.Validate.Struct(usr); err != nil {
		fmt.Fprintf(os.Stderr, "User validation failed: %v\n", err.(validator.ValidationErrors))
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	//generate salted pass hash
	if err := usr.HashPass(usr.Pass); err != nil {
		fmt.Fprintf(os.Stderr, "Can't hash the password: %v\n", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	//add user to db, get their id
	conn, err := a.p.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer conn.Release()
	row := conn.QueryRow(context.Background(),
		"INSERT INTO users(name, email, pass) VALUES ($1, $2, $3) RETURNING id", usr.Name, usr.Email, usr.Pass)

	//check if insert was successful
	var id int32
	if err = row.Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			http.Error(w, "", http.StatusConflict)
		} else if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.CheckViolation {
			http.Error(w, "", http.StatusBadRequest)
		} else {
			fmt.Fprintf(os.Stderr, "Unable to INSERT: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
		}
		return
	}

	//generate token pair
	tp, err := genTokenPair(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't gen token: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	//store token pair in redis
	err = storeTokenPair(a.rdb, id, tp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't store token: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	//marshal and return token pair
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{
		"access_token":  tp.at.token,
		"refresh_token": tp.rt.token,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
	}
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	uuid := gcontext.GetUUID(r.Context())

	deleted, err := a.rdb.Del(context.Background(), uuid).Result()
	if err != nil || deleted == 0 {
		fmt.Fprintf(os.Stderr, "Couldn't delete token from redis: %v", err)
		http.Error(w, "", http.StatusUnauthorized)
		return
	}
}

func (a *App) refresh(w http.ResponseWriter, r *http.Request) {
	uuid := gcontext.GetUUID(r.Context())

	deleted, err := a.rdb.Del(context.Background(), uuid).Result()
	if err != nil || deleted == 0 {
		fmt.Fprintf(os.Stderr, "Couldn't delete token from redis: %v", err)
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	userID := gcontext.GetUserID(r.Context())
	tp, err := genTokenPair(userID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't gen token: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	//store token pair in redis
	err = storeTokenPair(a.rdb, userID, tp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't store token: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	//marshal and return token pair
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(map[string]string{
		"access_token":  tp.at.token,
		"refresh_token": tp.rt.token,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
	}
}

func genTokenPair(userId int32) (*tokenPair, error) {
	tp := &tokenPair{}
	at, err := genToken(userId, time.Now().Add(time.Minute*15))
	if err != nil {
		return nil, err
	}
	tp.at = at

	rt, err := genToken(userId, time.Now().Add(time.Hour*24*7))
	if err != nil {
		return nil, err
	}
	tp.rt = rt

	return tp, nil
}

func genToken(userId int32, expires time.Time) (*tokenDetails, error) {
	td := &tokenDetails{}
	td.expires = expires.Unix()
	td.uuid = uuid.NewString()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"uuid":    td.uuid,
		"exp":     td.expires,
	})
	var err error
	td.token, err = token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return nil, err
	}

	return td, nil
}

func storeTokenPair(rdb *redis.Client, userid int32, tp *tokenPair) error {
	at := time.Unix(tp.at.expires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(tp.rt.expires, 0)
	now := time.Now()

	err := rdb.Set(context.Background(), tp.at.uuid, strconv.Itoa(int(userid)), at.Sub(now)).Err()
	if err != nil {
		return err
	}
	err = rdb.Set(context.Background(), tp.rt.uuid, strconv.Itoa(int(userid)), rt.Sub(now)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (a *App) withToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//extract jwt from header
		tokenString := r.Header.Get("Authorization")
		if len(tokenString) == 0 {
			fmt.Fprintf(os.Stderr, "No token")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}
		splitToken := strings.Split(tokenString, "Bearer ")
		if len(splitToken) < 2 {
			fmt.Fprintf(os.Stderr, "Invalid bearer token")
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		tokenString = splitToken[1]

		//parse & validate jwt
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid token: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}
		//put jwt in context
		ctx := gcontext.WithJwt(r.Context(), token)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (a *App) withClaims(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := gcontext.GetJwt(r.Context())
		if token == nil {
			fmt.Fprintf(os.Stderr, "Token didn't reach middleware: %v", token)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			fmt.Fprintf(os.Stderr, "Invalid token: %v", token)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		uuid, ok := claims["uuid"].(string)
		if !ok || uuid == "" {
			fmt.Fprintf(os.Stderr, "Couldn't extract uuid from token: %v", token)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		uidStr, err := a.rdb.Get(context.Background(), uuid).Result()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Access token invalid: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}
		uid, err := strconv.ParseInt(uidStr, 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't convert redis str to int32: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}
		usrCtx := gcontext.WithUserID(r.Context(), int32(uid))
		uuidCtx := gcontext.WithUUID(usrCtx, uuid)
		r = r.WithContext(uuidCtx)

		next.ServeHTTP(w, r)
	})
}
