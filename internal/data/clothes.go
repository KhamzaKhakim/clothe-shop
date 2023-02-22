package data

import (
	"clothing-store/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"strings"
	"time"
)

type Clothe struct {
	ID       int64    `json:"id"`
	Name     string   `json:"name"`
	Price    int64    `json:"price"`
	Brand    string   `json:"brand"`
	Color    string   `json:"color"`
	Sizes    []string `json:"sizes"`
	Sex      string   `json:"sex,omitempty"`
	Type     string   `json:"type,omitempty"`
	ImageURL string   `json:"image_url,omitempty"`
}

func ValidateClothe(v *validator.Validator, clothe *Clothe) {
	v.Check(clothe.Name != "", "name", "must be provided")
	v.Check(clothe.Brand != "", "brand", "must be provided")
	v.Check(clothe.Color != "", "color", "must be provided")
	v.Check(clothe.Sex != "", "sex", "must be provided")
	v.Check(clothe.Type != "", "type", "must be provided")
	v.Check(clothe.ImageURL != "", "image_url", "must be provided")
	v.Check(len(clothe.Name) <= 500, "name", "must not be more than 500 bytes long")
	v.Check(clothe.Price != 0, "price", "must be provided")
	v.Check(clothe.Price > 0, "price", "must be a positive integer")
	v.Check(clothe.Sizes != nil, "sizes", "must be provided")
	v.Check(len(clothe.Sizes) >= 1, "sizes", "must contain at least 1 size")
	v.Check(validator.Unique(clothe.Sizes), "sizes", "must not contain duplicate values")
}

type ClotheModel struct {
	DB *sql.DB
}

func (m ClotheModel) Insert(clothe *Clothe) error {
	query := `INSERT INTO clothes (name, price, brand, color, sizes, sex, type, image_url)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id`
	args := []any{clothe.Name, clothe.Price, clothe.Brand, clothe.Color, pq.Array(clothe.Sizes),
		clothe.Sex, clothe.Type, clothe.ImageURL}
	return m.DB.QueryRow(query, args...).Scan(&clothe.ID)
}

func (m ClotheModel) Get(id int64) (*Clothe, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `
		SELECT *
		FROM clothes
		WHERE id = $1`
	var clothe Clothe

	err := m.DB.QueryRow(query, id).Scan(
		&clothe.ID,
		&clothe.Name,
		&clothe.Price,
		&clothe.Brand,
		&clothe.Color,
		pq.Array(&clothe.Sizes),
		&clothe.Sex,
		&clothe.Type,
		&clothe.ImageURL,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &clothe, nil
}

func (m ClotheModel) Update(clothe *Clothe) error {
	query := `
			UPDATE clothes
			SET name = $1, price = $2, brand = $3, color = $4, sizes = $5, 
			    sex = $6, type = $7, image_url = $8 
			WHERE id = $9
			RETURNING id`
	args := []any{
		clothe.Name,
		clothe.Price,
		clothe.Brand,
		clothe.Color,
		pq.Array(clothe.Sizes),
		clothe.Sex,
		clothe.Type,
		clothe.ImageURL,
		clothe.ID,
	}
	return m.DB.QueryRow(query, args...).Scan(&clothe.ID)
}

func (m ClotheModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
				DELETE FROM clothes
				WHERE id = $1`
	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (m ClotheModel) GetAll(name string, brand string, priceMax int64, priceMin int64,
	sizes []string, color string, type_ string, sex string,
	filters Filters) ([]*Clothe, error) {
	sizesUpper := []string{}

	for i := 0; i < len(sizes); i++ {
		sizesUpper = append(sizesUpper, strings.ToUpper(sizes[i]))
	}
	println(brand)
	query := fmt.Sprintf(`
								SELECT *
								FROM clothes
								WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
								AND (clothes.sizes @> $2 OR $2 = '{}')
								AND price < $3 AND price > $4
								AND lower(brand) = lower($5)
								ORDER BY %s %s, id ASC LIMIT $6 OFFSET $7`, filters.sortColumn(), filters.sortDirection())

	if brand == "" {
		query = fmt.Sprintf(`
								SELECT *
								FROM clothes
								WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
								AND (clothes.sizes @> $2 OR $2 = '{}')
								AND price < $3 AND price > $4
								AND lower(brand) != lower($5)
								ORDER BY %s %s, id ASC LIMIT $6 OFFSET $7`, filters.sortColumn(), filters.sortDirection())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, name, pq.Array(sizesUpper), priceMax, priceMin, brand, filters.limit(),
		filters.offset())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clothes := []*Clothe{}
	for rows.Next() {
		var clothe Clothe

		err := rows.Scan(
			&clothe.ID,
			&clothe.Name,
			&clothe.Price,
			&clothe.Brand,
			&clothe.Color,
			pq.Array(&clothe.Sizes),
			&clothe.Sex,
			&clothe.Type,
			&clothe.ImageURL,
		)
		if err != nil {
			return nil, err
		}
		clothes = append(clothes, &clothe)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return clothes, nil
}
