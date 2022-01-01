package memRepos

import (
	Cerr "mmr/errors"
	"mmr/models"
	"sync"
)

type Category struct {
	storage map[int32]models.Category
	mu      sync.Mutex
}

func NewCategory(storage map[int32]models.Category) *Category {
	return &Category{
		storage: storage,
		mu:      sync.Mutex{},
	}
}

func (ctg *Category) List() ([]models.Category, Cerr.CError) {
	ctg.mu.Lock()
	defer ctg.mu.Unlock()

	categories := make([]models.Category, len(ctg.storage))
	for _, category := range ctg.storage {
		categories = append(categories, category)
	}

	return categories, nil
}

func (ctg *Category) Get(id int32) (*models.Category, Cerr.CError) {
	ctg.mu.Lock()
	defer ctg.mu.Unlock()

	category, ok := ctg.storage[id]
	if !ok {
		return nil, Cerr.NewNotFound("category id")
	}

	return &category, nil
}
