package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Clothes     ClotheModel
	Users       UserModel
	Brands      BrandModel
	Tokens      TokenModel
	Permissions PermissionModel
	Roles       RolesModel
	Carts       CartsModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Clothes:     ClotheModel{DB: db},
		Users:       UserModel{DB: db},
		Brands:      BrandModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Permissions: PermissionModel{DB: db},
		Roles:       RolesModel{DB: db},
		Carts:       CartsModel{DB: db},
	}
}
