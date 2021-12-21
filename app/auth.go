package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"mmr/models"
	"mmr/shared"
	"net/http"
	"os"
	"time"
)

func (a *App) SignIn(w http.ResponseWriter, r *http.Request) {
	//serialize user
	var usr models.User
	if err := json.NewDecoder(r.Body).Decode(&usr); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid request: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//validate user
	if err := shared.Validate.Struct(usr); err != nil {
		fmt.Fprintf(os.Stderr, "User validation failed: %v\n", err.(validator.ValidationErrors))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//find the user in db
	conn, err := a.p.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Release()
	row := conn.QueryRow(context.Background(),
		"SELECT id, name, email, pass FROM users WHERE email = $1", usr.Email)

	//serialize db user or return 401 if they don't exist
	var dbUsr models.User
	if err = row.Scan(&dbUsr.Id, &dbUsr.Name, &dbUsr.Email, &dbUsr.Pass); err == pgx.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to SELECT: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//validate password
	if err = dbUsr.ValidatePass(usr.Pass); err != nil {
		fmt.Fprintf(os.Stderr, "Pass is not correct: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//generate and sign jwt
	tokenString, err := genToken(dbUsr.Id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't sign token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//marshal and return jwt
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err = json.NewEncoder(w).Encode(tokenString); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (a *App) SignUp(w http.ResponseWriter, r *http.Request) {
	//serialize user
	var usr models.User
	if err := json.NewDecoder(r.Body).Decode(&usr); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid request: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//validate user
	if err := shared.Validate.Struct(usr); err != nil {
		fmt.Fprintf(os.Stderr, "User validation failed: %v\n", err.(validator.ValidationErrors))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//generate salted pass hash
	if err := usr.HashPass(usr.Pass); err != nil {
		fmt.Fprintf(os.Stderr, "Can't hash the password: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//add user to db, get their id
	conn, err := a.p.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
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
			w.WriteHeader(http.StatusConflict)
		} else if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.CheckViolation {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			fmt.Fprintf(os.Stderr, "Unable to INSERT: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	tokenString, err := genToken(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't sign token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//marshal and return jwt
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(tokenString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
	}
}

func genToken(id int32) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"exp":     time.Now().Add(time.Minute * 15).Unix(),
	})
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
