package app

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"mmr/services"
	postgresql "mmr/services/repositories"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func svcF() *services.User {
	//init service
	p, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	repo := postgresql.NewUser(p)
	svc := services.NewUser(repo)

	return svc
}

func rdbF() *redis.Client {
	//init redis
	dsn := os.Getenv("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: dsn,
	})
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Unable to connect to redis:", err)
	}

	return rdb
}

var a = NewApp(svcF(), rdbF())

func TestRegister(t *testing.T) {
	tt := []struct {
		name       string
		method     string
		user       *testUser
		willAdd    bool
		statusCode int
	}{
		{
			name:   "with valid data",
			method: http.MethodPost,
			user: &testUser{
				Name:  "test_user",
				Email: "test_email@test.com",
				Pass:  "123456",
			},
			willAdd:    true,
			statusCode: http.StatusOK,
		},
		{
			name:       "with duplicate email",
			method:     http.MethodPost,
			user:       &validUser,
			willAdd:    false,
			statusCode: http.StatusConflict,
		},
		{
			name:   "with no email",
			method: http.MethodPost,
			user: &testUser{
				Name: "test_user2",
				Pass: "123456",
			},
			willAdd:    false,
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "with invalid email",
			method: http.MethodPost,
			user: &testUser{
				Name:  "test_user2",
				Email: "invalid_email",
				Pass:  "123456",
			},
			willAdd:    false,
			statusCode: http.StatusBadRequest,
		},
	}

	_db := &db{}
	_db.init(t)
	t.Cleanup(_db.clearDb)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			conn, release := _db.getConn()
			defer release()

			usersBefore := countUsers(conn, t)

			body, err := json.Marshal(tc.user)
			if err != nil {
				t.Fatal("Can't marshal the user", err)
			}

			request := httptest.NewRequest(tc.method, "/auth/register", bytes.NewReader(body))
			responseRecorder := httptest.NewRecorder()

			a.r.ServeHTTP(responseRecorder, request)
			defer conn.Exec(context.Background(), "DELETE FROM users WHERE email = $1", tc.user.Email)

			//Status is correct
			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			//token is returned
			resBody := make(map[string]string)
			json.Unmarshal(responseRecorder.Body.Bytes(), &resBody)
			if tc.willAdd && (len(resBody["access_token"]) == 0 || len(resBody["refresh_token"]) == 0) {
				t.Error("token was not returned: ", resBody)
			}

			//User was added to DB
			usersAfter := countUsers(conn, t) //TODO: another test goroutine could insert a user before this call
			if (tc.willAdd && usersBefore+1 != usersAfter) ||
				(!tc.willAdd && usersBefore != usersAfter) {
				t.Errorf("Unexpected user count. Users before '%d', users after '%d'", usersBefore, usersAfter)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	tt := []struct {
		name       string
		method     string
		user       *testUser
		needsToken bool
		statusCode int
	}{
		{
			name:       "with valid credentials",
			method:     http.MethodPost,
			user:       &validUser,
			needsToken: true,
			statusCode: http.StatusOK,
		},
		{
			name:   "with invalid credentials",
			method: http.MethodPost,
			user: &testUser{
				Email: "invalid@test.com",
				Pass:  "123456",
			},
			needsToken: false,
			statusCode: http.StatusNotFound,
		},
		{
			name:   "with missing password",
			method: http.MethodPost,
			user: &testUser{
				Email: "test_email@test.com",
			},
			needsToken: false,
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "with missing email",
			method: http.MethodPost,
			user: &testUser{
				Pass: "123456",
			},
			needsToken: false,
			statusCode: http.StatusBadRequest,
		},
	}

	_db := &db{}
	_db.init(t)
	t.Cleanup(_db.clearDb)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.user)
			if err != nil {
				t.Fatal("Can't marshal the user", err)
			}

			request := httptest.NewRequest(tc.method, "/auth/login", bytes.NewReader(body))
			responseRecorder := httptest.NewRecorder()

			a.r.ServeHTTP(responseRecorder, request)

			//Status is correct
			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			//token is returned
			resBody := make(map[string]string)
			json.Unmarshal(responseRecorder.Body.Bytes(), &resBody)
			if tc.needsToken && (len(resBody["access_token"]) == 0 || len(resBody["refresh_token"]) == 0) {
				t.Error("token was not returned: ", resBody)
			}
		})
	}
}

func countUsers(conn *pgxpool.Conn, t *testing.T) int {
	row := conn.QueryRow(context.Background(),
		"SELECT COUNT(id) FROM users")
	var usrCount int
	if err := row.Scan(&usrCount); err != nil {
		t.Fatal("Unable to get user count from db", err)
	}

	return usrCount
}
