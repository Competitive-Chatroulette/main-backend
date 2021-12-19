package handlers

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
	"github.com/jackc/pgx/v4/pgxpool"
	"mmr/app/models"
	"mmr/shared"
	"net/http"
	"os"
	"time"
)

func SignIn(p *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	//serialize user
	var usr models.User
	err := json.NewDecoder(r.Body).Decode(&usr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid request: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//validate user
	err = shared.Validate.Struct(usr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "User validation failed: %v\n", err.(validator.ValidationErrors))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//find the user in db
	conn, err := p.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Release()
	row := conn.QueryRow(context.Background(),
		"SELECT id, name, email, pass FROM users WHERE email = $1", usr.Email)

	//serialize db user or return 401 if they don't exist
	var resUsr models.User
	err = row.Scan(&resUsr.Id, &resUsr.Name, &resUsr.Email, &resUsr.Pass)
	if err == pgx.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to SELECT: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//validate password
	err = resUsr.ValidatePass(usr.Pass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Pass is not correct: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resUsr.Pass = "" //remove password hash from the user object we are returning

	//generate and sign jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": resUsr.Id,
		"nbf":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't sign token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//marshal and return jwt token
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(tokenString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func SignUp(p *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	//serialize user
	var usr models.User
	err := json.NewDecoder(r.Body).Decode(&usr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid request: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//validate user
	err = shared.Validate.Struct(usr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "User validation failed: %v\n", err.(validator.ValidationErrors))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//generate salted pass hash
	err = usr.HashPass(usr.Pass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't hash the password: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//add user to db, get their id
	conn, err := p.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Release()
	row := conn.QueryRow(context.Background(),
		"INSERT INTO users(name, email, pass) VALUES ($1, $2, $3) RETURNING id", usr.Name, usr.Email, usr.Pass)

	var id int
	err = row.Scan(&id)
	if err != nil {
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

	//generate and sign jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"nbf":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't sign token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//marshal and return jwt token
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(tokenString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
	}
}
