package main

import (
	"database/sql"
	api "wishlist-backend/controllers"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"

	// _ "modernc.org/sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	sqlite, err := sql.Open("sqlite3", "file:test.db?_foreign_keys=on")

	db := goqu.New("sqlite3", sqlite)

	if err != nil {
		return
	}

	api.New(db).Run("localhost:8000")
}
