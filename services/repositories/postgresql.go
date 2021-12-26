package postgresql

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	cerr "mmr/errors"
	"mmr/models"
	"os"
)

type User struct {
	p *pgxpool.Pool
}

func NewUser(p *pgxpool.Pool) *User {
	return &User{
		p: p,
	}
}

func (usr *User) Create(user *models.User) (int32, cerr.CError) {
	conn, err := usr.p.Acquire(context.Background())
	if err != nil {
		return 0, cerr.NewInternal()
	}
	defer conn.Release()

	row := conn.QueryRow(context.Background(),
		"INSERT INTO users(name, email, pass) VALUES ($1, $2, $3) RETURNING id", user.Name, user.Email, user.Pass)
	var userID int32
	if err = row.Scan(&userID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, cerr.NewNotUnique(pgErr.ColumnName)
		} else {
			fmt.Fprintf(os.Stderr, "Unable to INSERT: %v", err)
			return 0, cerr.NewInternal()
		}
	}

	return userID, nil
}

func (usr *User) FindById(userID int32) (*models.User, cerr.CError) {
	conn, err := usr.p.Acquire(context.Background())
	if err != nil {
		return nil, cerr.NewInternal()
	}
	defer conn.Release()

	row := conn.QueryRow(context.Background(),
		"SELECT id, name, email FROM users WHERE id = $1", userID)
	var dbUsr *models.User
	if err = row.Scan(dbUsr.Id, dbUsr.Name, dbUsr.Email); err == pgx.ErrNoRows {
		return nil, cerr.NewNotFound("user")
	} else if err != nil {
		return nil, cerr.NewInternal()
	}

	return dbUsr, nil
}

func (usr *User) FindByEmail(email string) (*models.User, cerr.CError) {
	conn, err := usr.p.Acquire(context.Background())
	if err != nil {
		return nil, cerr.NewInternal()
	}
	defer conn.Release()

	row := conn.QueryRow(context.Background(),
		"SELECT id, name, email, pass FROM users WHERE email = $1", email)
	var dbUsr *models.User
	if err = row.Scan(dbUsr.Id, dbUsr.Name, dbUsr.Email, dbUsr.Pass); err == pgx.ErrNoRows {
		return nil, cerr.NewNotFound("email")
	} else if err != nil {
		return nil, cerr.NewInternal()
	}

	return dbUsr, nil
}
