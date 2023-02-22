package data

import (
	"clothing-store/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

type Brand struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Country     string `json:"country"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url,omitempty"`
}

func ValidateBrand(v *validator.Validator, brand *Brand) {
	v.Check(brand.Name != "", "name", "must be provided")
	v.Check(brand.Country != "", "country", "must be provided")
	v.Check(brand.Description != "", "description", "must be provided")
	v.Check(brand.ImageURL != "", "image_url", "must be provided")
}

type BrandModel struct {
	DB *sql.DB
}

func (m BrandModel) Insert(brand *Brand) error {
	query := `INSERT INTO brands (name, country, description, image_url)
				VALUES ($1, $2, $3, $4)
				RETURNING id`
	args := []any{brand.Name, brand.Country, brand.Description, brand.ImageURL}
	return m.DB.QueryRow(query, args...).Scan(&brand.ID)
}

func (m BrandModel) Get(id int64) (*Brand, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `
		SELECT *
		FROM brands
		WHERE id = $1`
	var brand Brand

	err := m.DB.QueryRow(query, id).Scan(
		&brand.ID,
		&brand.Name,
		&brand.Country,
		&brand.Description,
		&brand.ImageURL,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &brand, nil
}

func (m BrandModel) Update(brand *Brand) error {
	query := `
			UPDATE brands
			SET name = $1, country = $2, description = $3, image_url = $4
			WHERE id = $5
			RETURNING id`
	args := []any{
		brand.Name,
		brand.Country,
		brand.Description,
		brand.ImageURL,
		brand.ID,
	}
	return m.DB.QueryRow(query, args...).Scan(&brand.ID)
}

func (m BrandModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
				DELETE FROM brands
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

func (m BrandModel) GetAll() ([]*Brand, error) {
	query := fmt.Sprintf(`
								SELECT *
								FROM brands ORDER BY id`)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	brands := []*Brand{}
	for rows.Next() {
		var brand Brand

		err := rows.Scan(
			&brand.ID,
			&brand.Name,
			&brand.Country,
			&brand.Description,
			&brand.ImageURL,
		)
		if err != nil {
			return nil, err
		}
		brands = append(brands, &brand)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return brands, nil
}

func (m BrandModel) GetAllBrandNames() (names []string) {
	sel := "SELECT name FROM brands"

	rows, err := m.DB.Query(sel)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		names = append(names, name)
	}
	return
}
