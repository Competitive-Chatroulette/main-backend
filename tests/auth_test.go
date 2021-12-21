package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v4/pgxpool"
	"mmr/app"
	"net/http"
	"net/http/httptest"
	"testing"
)

var a = app.NewApp()

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

			request := httptest.NewRequest(tc.method, "/auth/signup", bytes.NewReader(body))
			responseRecorder := httptest.NewRecorder()

			a.SignUp(responseRecorder, request)
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
	_db.init(t)
	t.Cleanup(_db.clearDb)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.user)
			if err != nil {
				t.Fatal("Can't marshal the user", err)
			}

			request := httptest.NewRequest(tc.method, "/auth/signin", bytes.NewReader(body))
			responseRecorder := httptest.NewRecorder()

			a.SignIn(responseRecorder, request)

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
