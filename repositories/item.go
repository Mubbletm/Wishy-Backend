package repository

import (
	"github.com/doug-martin/goqu/v9"
)

type Item struct {
	Model[string]
	WishlistId  string `json:"-" db:"WishlistId"`
	Url         string `json:"url" db:"Url"`
	Name        string `json:"name" db:"Name"`
	Description string `json:"description" db:"Description"`
	Image       string `json:"image" db:"Image"`
}

type ItemRepository struct {
	*AbstractSQLiteRepository[Item, string]
}

func NewItemRepository(db *goqu.Database) *ItemRepository {
	repo := &ItemRepository{
		&AbstractSQLiteRepository[Item, string]{
			db:     db,
			dbName: "Item",
			empty:  Item{},
		},
	}
	return repo
}

func (repo *ItemRepository) RemoveId(item *Item) {
	item.Id = ""
}
