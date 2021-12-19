package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"mmr/app/handlers"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type testUser struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Pass  string `json:"pass"`
}

var validUser = testUser{
	Name:  "cool guy",
	Email: "valid_test@test.com",
	Pass:  "123456",
}

type db struct {
	P *pgxpool.Pool
	t *testing.T
}

func (_db *db) Init(t *testing.T) {
	_db.t = t
	_db.initP()
	_db.fillDb()
}

func (_db *db) initP() {
	p, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		_db.t.Fatal("Unable to connect to database:", err)
	}

	_db.P = p
}

func (_db *db) GetConn() (*pgxpool.Conn, func()) {
	conn, err := _db.P.Acquire(context.Background())
	if err != nil {
		_db.t.Fatal(err)
	}
	return conn, conn.Release
}

func (_db *db) fillDb() {
	conn, release := _db.GetConn()
	defer release()

	pass := hashPass(validUser.Pass, _db.t)
	_, err := conn.Exec(context.Background(), "INSERT INTO users(name, email, pass) VALUES($1, $2, $3)", validUser.Name, validUser.Email, pass)
	if err != nil {
		_db.t.Fatal("Could not insert user", err)
	}
}

func (_db *db) ClearDb() {
	conn, release := _db.GetConn()
	defer release()

	_, err := conn.Exec(context.Background(), "DELETE FROM users WHERE email=$1", validUser.Email)
	if err != nil {
		_db.t.Fatal("Could not cleanup user", err)
	}
}

func hashPass(pass string, t *testing.T) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal("Can't hash the password: ", err)
	}

	return string(hash)
}

func TestSignup(t *testing.T) {
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
			statusCode: http.StatusCreated,
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
	_db.Init(t)
	t.Cleanup(_db.ClearDb)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			conn, release := _db.GetConn()
			defer release()

			usersBefore := countUsers(conn, t)

			body, err := json.Marshal(tc.user)
			if err != nil {
				t.Fatal("Can't marshal the user", err)
			}

			request := httptest.NewRequest(tc.method, "/auth/signup", bytes.NewReader(body))
			responseRecorder := httptest.NewRecorder()

			handlers.SignUp(_db.P, responseRecorder, request)
			defer conn.Exec(context.Background(), "DELETE FROM users WHERE email = $1", tc.user.Email)

			//Status is correct
			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			//Token is returned
			if tc.willAdd && responseRecorder.Body.Len() == 0 {
				t.Errorf("Token was not returned")
			}

			//User was added to DB
			usersAfter := countUsers(conn, t)
			if (tc.willAdd && usersBefore+1 != usersAfter) ||
				(!tc.willAdd && usersBefore != usersAfter) {
				t.Errorf("Unexpected user count. Users before '%d', users after '%d'", usersBefore, usersAfter)
			}
		})
	}
}

func TestSignin(t *testing.T) {
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
			statusCode: http.StatusUnauthorized,
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
	_db.Init(t)
	t.Cleanup(_db.ClearDb)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.user)
			if err != nil {
				t.Fatal("Can't marshal the user", err)
			}

			request := httptest.NewRequest(tc.method, "/auth/signin", bytes.NewReader(body))
			responseRecorder := httptest.NewRecorder()

			handlers.SignIn(_db.P, responseRecorder, request)

			//Status is correct
			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			//Token is returned
			if tc.needsToken && responseRecorder.Body.Len() == 0 {
				t.Errorf("Token was not returned")
			}
		})
	}
}

func countUsers(conn *pgxpool.Conn, t *testing.T) int {
	row := conn.QueryRow(context.Background(),
		"SELECT COUNT(id) FROM users")
	var usrCount int
	err := row.Scan(&usrCount)
	if err != nil {
		t.Fatal("Unable to get user count from db", err)
	}

	return usrCount
}
