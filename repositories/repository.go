package repository

import (
	"os"

	"github.com/doug-martin/goqu/v9"
)

type Model[I any] struct {
	Id I `json:"id" db:"Id" goqu:"skipinsert"`
}

type Repository[T interface{}, I any] interface {
	Update(o T, id *I) error
	Add(o T) (*T, error)
	GetById(id I) (*T, error)
	GetAll() (*[]T, error)
	DeleteById(id T) error
	RemoveId(*T)
}

// An abstract repository that partially implements the Repository[T, I] interface.
//
// The methods `Add(o T) (*T, error)` and `Update(o T, id *I) error` do not have a proper implementation
// in this abstract repository. They have method placeholders that will return errors when invoked.
//
// Generic[T] is the type of the model of this controller.
//
// Generic[I] is the type of the model's ID of this controller. The ID of your model
// should always be named 'Id' in the database for it to work with this controller and its
// repository.
type AbstractSQLiteRepository[T interface{}, I any] struct {
	Repository[T, I]
	db     *goqu.Database
	empty  T
	dbName string
}

// Searches its database for the model with the given ID.
func (repo *AbstractSQLiteRepository[T, I]) GetById(id I) (*T, error) {
	template := repo.empty
	found, err := repo.db.From(repo.dbName).Where(goqu.Ex{
		"Id": id,
	}).ScanStruct(&template)

	if !found {
		return nil, os.ErrNotExist
	}

	if err != nil {
		return nil, err
	}

	return &template, nil
}

// Returns all values in its model's database table.
func (repo *AbstractSQLiteRepository[T, I]) GetAll() (*[]T, error) {
	models := []T{}
	err := repo.db.From(repo.dbName).ScanStructs(&models)

	if err != nil {
		return nil, err
	}

	return &models, nil
}

// Updates a given model's value. If an ID is provided as the second value, it should take precedent over any IDs in the model.
func (repo *AbstractSQLiteRepository[T, I]) Update(o T, id *I) error {
	_, err := repo.db.Update(repo.dbName).Set(o).
		Where(goqu.C("Id").Eq(id)).
		Executor().Exec()

	return err
}

// Adds a value to the model's database table. IDs will be auto generated, provided IDs should be ignored.
func (repo *AbstractSQLiteRepository[T, I]) Add(o T) (*T, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}

	result, err := tx.Insert(repo.dbName).Rows(o).Executor().Exec()

	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return nil, rErr
		}
		return nil, err
	}

	rowId, err := result.LastInsertId()

	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return nil, rErr
		}
		return nil, err
	}

	template := repo.empty
	_, err = tx.From(repo.dbName).
		Where(goqu.C("rowid").Eq(rowId)).
		ScanStruct(&template)

	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return nil, rErr
		}
		return nil, err
	}

	tx.Commit()
	return &template, nil
}

func (repo *AbstractSQLiteRepository[T, I]) DeleteById(id T) error {
	result, err := repo.db.Delete(repo.dbName).Where(goqu.C("Id").Eq(id)).Executor().Exec()

	if err != nil {
		return err
	}

	if count, err := result.RowsAffected(); count <= 0 && err != nil {
		return os.ErrNotExist
	}

	return nil
}
