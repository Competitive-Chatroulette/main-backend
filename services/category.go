package services

import (
	cerr "mmr/errors"
	"mmr/models"
)

type CategoryRepository interface {
	List() ([]models.Category, cerr.CError)
	Get(id int32) (*models.Category, cerr.CError)
}

type Category struct {
	repo CategoryRepository
}

func NewCategory(repo CategoryRepository) *Category {
	return &Category{
		repo: repo,
	}
}

func (ctg *Category) List() ([]models.Category, cerr.CError) {
	return ctg.repo.List()
}

func (ctg *Category) Get(id int32) (*models.Category, cerr.CError) {
	return ctg.repo.Get(id)
}
