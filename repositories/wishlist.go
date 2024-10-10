package repository

import (
	"errors"
	"os"

	"github.com/doug-martin/goqu/v9"
)

type Wishlist struct {
	Model[string]
	Name      string `json:"name" db:"Name"`
	Password  string `json:"-" db:"Password" goqu:"skipinsert"`
	Ownership string `json:"-" db:"Ownership"`
}

type WishlistBody struct {
	Model[string]
	Name      string `json:"name"`
	Ownership string `json:"ownership"`
}

type UnlockedWishlist struct {
	Model[string]
	Name      string `json:"name" db:"Name"`
	Password  string `json:"password" db:"Password" goqu:"skipinsert"`
	Ownership string `json:"ownership" db:"Ownership"`
}

type WishlistRepository struct {
	*AbstractSQLiteRepository[Wishlist, string]
}

type WishlistViewer struct {
	WishlistId  string   `json:"-" db:"WishlistId"`
	Wishlist    Wishlist `json:"wishlist" db:"-"`
	Permissions string   `json:"permission" db:"Permissions"`
	Ownership   string   `json:"-" db:"Ownership"`
}

type WishlistPermissioned struct {
	Wishlist
	Permissions string `json:"permission" db:"Permissions"`
}

func NewWishlistRepository(db *goqu.Database) *WishlistRepository {
	repo := &WishlistRepository{
		&AbstractSQLiteRepository[Wishlist, string]{
			db:     db,
			dbName: "Wishlist",
			empty:  Wishlist{},
		},
	}
	return repo
}

func (repo *WishlistRepository) GetItems(id string) ([]Item, error) {
	var items []Item
	err := repo.db.From("Item").Where(goqu.C("WishlistId").Eq(id)).ScanStructs(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (repo *WishlistRepository) GetOwnedWishlists(ownership string) ([]Wishlist, error) {
	var wishlists []Wishlist
	err := repo.db.From("Wishlist").Where(goqu.C("Ownership").Eq(ownership)).ScanStructs(&wishlists)

	if err != nil {
		return nil, err
	}
	return wishlists, nil
}

func (repo *WishlistRepository) GetSavedWishlists(ownership string) ([]WishlistPermissioned, error) {
	var wishlists []WishlistPermissioned

	err := repo.db.From("Wishlist").Select("Wishlist.*", "WishlistViewer.Permissions").InnerJoin(
		goqu.T("WishlistViewer"),
		goqu.On(goqu.Ex{
			"WishlistViewer.WishlistId": goqu.I("Wishlist.Id"),
		}),
	).Where(goqu.I("WishlistViewer.Ownership").Eq(ownership)).ScanStructs(&wishlists)

	if err != nil {
		return nil, err
	}
	return wishlists, nil
}

func (repo *WishlistRepository) GetPermission(wishlistId string, ownership string) (string, error) {
	result, err := repo.db.From("WishlistViewer").Select(goqu.C("Permissions")).Where(goqu.And(
		goqu.C("WishlistId").Eq(wishlistId),
		goqu.C("Ownership").Eq(ownership),
	)).Executor().Query()

	if err != nil {
		return "", err
	}

	if !result.Next() {
		return "", os.ErrNotExist
	}

	var permission string
	result.Scan(&permission)
	return permission, nil
}

func (repo *WishlistRepository) RegisterPermission(wishlistId string, ownership string, password string) error {
	var model goqu.Record
	// If password is present, verify password and change permission from viewer to editor.
	if len(password) > 0 {
		res, err := repo.db.From("Wishlist").Where(goqu.And(
			goqu.C("Id").Eq(wishlistId),
			goqu.C("Password").Eq(password),
		)).Executor().Query()

		if err != nil {
			return err
		}

		// When no results, return error.
		if !res.Next() {
			return os.ErrNotExist
		}

		// Prepare data for insertion/update.
		model = goqu.Record{
			"WishlistId":  wishlistId,
			"Permissions": "EDIT",
			"Ownership":   ownership,
		}
	} else {
		// Leave out permission because the database will default to viewer.
		// This way the user will stay an editor if an update happens instead
		// of an insert on an existing record with edit permissions.
		model = goqu.Record{
			"WishlistId": wishlistId,
			"Ownership":  ownership,
		}
	}

	// Check if the user already has permissions for the given wishlist.
	_, err := repo.GetPermission(wishlistId, ownership)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	// If the user does not have permissions, insert row, otherwise update row.
	if errors.Is(err, os.ErrNotExist) {
		_, err = repo.db.From("WishlistViewer").Insert().Rows(model).Executor().Exec()
	} else if len(password) > 0 {
		_, err = repo.db.From("WishlistViewer").Update().Where(goqu.And(
			goqu.C("WishlistId").Eq(wishlistId),
			goqu.C("Ownership").Eq(ownership),
		)).Set(model).Executor().Exec()
	}

	if err != nil {
		return err
	}
	return nil
}

func (repo *WishlistRepository) RemoveId(wishlist *Wishlist) {
	wishlist.Id = ""
}
