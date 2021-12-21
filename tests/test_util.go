package tests

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
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

func (_db *db) init(t *testing.T) {
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

func (_db *db) getConn() (*pgxpool.Conn, func()) {
	conn, err := _db.P.Acquire(context.Background())
	if err != nil {
		_db.t.Fatal(err)
	}
	return conn, conn.Release
}

func (_db *db) fillDb() {
	conn, release := _db.getConn()
	defer release()

	pass := hashPass(validUser.Pass, _db.t)
	_, err := conn.Exec(context.Background(), "INSERT INTO users(name, email, pass) VALUES($1, $2, $3)", validUser.Name, validUser.Email, pass)
	if err != nil {
		_db.t.Fatal("Could not insert user", err)
	}
}

func (_db *db) clearDb() {
	conn, release := _db.getConn()
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
