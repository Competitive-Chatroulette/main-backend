package postgresql

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	cerr "mmr/errors"
	"mmr/models"
	"os"
)

type Category struct {
	p *pgxpool.Pool
}

func NewCategory(p *pgxpool.Pool) *Category {
	return &Category{
		p: p,
	}
}

func (ctg *Category) List() ([]models.Category, cerr.CError) {
	conn, err := ctg.p.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		return nil, cerr.NewInternal()
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), "SELECT id, name FROM categories")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to SELECT categories: %v\n", err)
		return nil, cerr.NewInternal()
	}

	//skip rows with errors while scanning
	categories := make([]models.Category, 0)
	for rows.Next() {
		var category models.Category
		if err = rows.Scan(&category.Id, &category.Name); err == nil {
			categories = append(categories, category)
		}
	}

	//if none of the categories were scanned, return 500; else return what was successfully scanned
	if err := rows.Err(); err != nil {
		//log scanning errors
		fmt.Fprintf(os.Stderr, "Error while reading categories table: %s", err)

		if len(categories) == 0 {
			return nil, cerr.NewInternal()
		}
	}

	return categories, nil
}

func (ctg *Category) Get(id int32) (*models.Category, cerr.CError) {
	conn, err := ctg.p.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		return nil, cerr.NewInternal()
	}
	defer conn.Release()

	row := conn.QueryRow(context.Background(),
		"SELECT id, name FROM categories WHERE id = $1", id)

	var category models.Category
	if err = row.Scan(&category.Id, &category.Name); err == pgx.ErrNoRows {
		return nil, cerr.NewNotFound("category")
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to SELECT category: %v", err)
		return nil, cerr.NewInternal()
	}

	return &category, nil
}
