package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"time"
)

type Cart struct {
	UserID int64
	ID     []int64
}

type CartsModel struct {
	DB *sql.DB
}

func (m CartsModel) AddClotheForCart(userID int64, clothe Clothe) error {
	query := `UPDATE carts
	SET clothes_id = array_append(clothes_id, $1)
	WHERE user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, clothe.ID, userID)
	return err
}

func (m CartsModel) CreateCartForUser(userID int64) error {
	query := `
INSERT INTO carts VALUES ($1)`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, userID)
	return err
}

func (m CartsModel) GetById(id int64) ([]int64, error) {
	query := `
				SELECT clothes_id
				FROM carts
				WHERE user_id = $1`
	var clothes []int64
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(pq.Array(&clothes))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return clothes, nil
}
