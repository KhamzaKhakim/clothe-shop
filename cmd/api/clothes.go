package main

import (
	"clothing-store/internal/data"
	"clothing-store/internal/validator"
	"errors"
	"fmt"
	"net/http"
)

func (app *application) createClotheHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name     string   `json:"name"`
		Price    int64    `json:"price"`
		Brand    string   `json:"brand"`
		Color    string   `json:"color"`
		Sizes    []string `json:"sizes"`
		Sex      string   `json:"sex"`
		Type     string   `json:"type"`
		ImageURL string   `json:"image_url"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	clothe := &data.Clothe{
		Name:     input.Name,
		Price:    input.Price,
		Brand:    input.Brand,
		Color:    input.Color,
		Sizes:    input.Sizes,
		Sex:      input.Sex,
		Type:     input.Type,
		ImageURL: input.ImageURL,
	}
	v := validator.New()

	if data.ValidateClothe(v, clothe); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Clothes.Insert(clothe)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/clothes/%d", clothe.ID))
	err = app.writeJSON(w, http.StatusCreated, clothe, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showClotheHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	clothe, err := app.models.Clothes.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Encode the struct to JSON and send it as the HTTP response.
	err = app.writeJSON(w, http.StatusOK, clothe, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateClotheHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	clothe, err := app.models.Clothes.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	var input struct {
		Name     *string  `json:"name"`
		Price    *int64   `json:"price"`
		Brand    *string  `json:"brand"`
		Color    *string  `json:"color"`
		Sizes    []string `json:"sizes"`
		Sex      *string  `json:"sex"`
		Type     *string  `json:"type"`
		ImageURL *string  `json:"image_url"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		clothe.Name = *input.Name
	}
	if input.Price != nil {
		clothe.Price = *input.Price
	}
	if input.Brand != nil {
		clothe.Brand = *input.Brand
	}
	if input.Color != nil {
		clothe.Color = *input.Color
	}
	if input.Sizes != nil {
		clothe.Sizes = input.Sizes
	}
	if input.Sex != nil {
		clothe.Sex = *input.Sex
	}
	if input.Type != nil {
		clothe.Type = *input.Type
	}
	if input.ImageURL != nil {
		clothe.ImageURL = *input.ImageURL
	}

	v := validator.New()
	if data.ValidateClothe(v, clothe); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Clothes.Update(clothe)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, clothe, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteClotheHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Clothes.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "clothe successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listClothesHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name     string
		Brand    string
		PriceMax int64
		PriceMin int64
		Sizes    []string
		Color    string
		Type     string
		Sex      string
		data.Filters
	}
	v := validator.New()
	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")
	input.Brand = app.readString(qs, "brand", "")
	input.PriceMax = app.readInt(qs, "price_max", 10000000, v)
	input.PriceMin = app.readInt(qs, "price_min", 0, v)
	input.Name = app.readString(qs, "name", "")
	input.Sizes = app.readCSV(qs, "sizes", []string{})
	input.Sex = app.readString(qs, "sex", "")
	input.Sex = app.readString(qs, "sex", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = 20
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "price", "sex", "brand", "-id", "-name", "-price", "-sex", "-brand"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	allBrands := app.models.Brands.GetAllBrandNames()
	allBrands = append(allBrands, "")

	keys := data.Keys{
		PriceMax:       input.PriceMax,
		PriceMin:       input.PriceMin,
		Brand:          input.Brand,
		Sizes:          input.Sizes,
		SizesSafelist:  []string{"XS", "S", "M", "L", "XL", ""},
		BrandsSafelist: allBrands,
	}

	if data.ValidateKeys(v, keys); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	clothes, err := app.models.Clothes.GetAll(input.Name, input.Brand, input.PriceMax, input.PriceMin,
		input.Sizes, input.Color, input.Type, input.Sex, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, clothes, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
