package main

import (
	"clothing-store/internal/data"
	"clothing-store/internal/validator"
	"errors"
	"fmt"
	"net/http"
)

func (app *application) createBrandHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name        string `json:"name"`
		Country     string `json:"country"`
		Description string `json:"description"`
		ImageURL    string `json:"image_url"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	brand := &data.Brand{
		Name:        input.Name,
		Country:     input.Country,
		Description: input.Description,
		ImageURL:    input.ImageURL,
	}
	v := validator.New()

	if data.ValidateBrand(v, brand); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Brands.Insert(brand)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/brands/%d", brand.ID))
	err = app.writeJSON(w, http.StatusCreated, brand, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showBrandHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	brand, err := app.models.Brands.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, brand, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateBrandHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	brand, err := app.models.Brands.Get(id)
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
		Name        *string `json:"name"`
		Country     *string `json:"country"`
		Description *string `json:"description"`
		ImageURL    *string `json:"image_url"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		brand.Name = *input.Name
	}
	if input.Country != nil {
		brand.Country = *input.Country
	}
	if input.Description != nil {
		brand.Description = *input.Description
	}
	if input.ImageURL != nil {
		brand.ImageURL = *input.ImageURL
	}

	v := validator.New()
	if data.ValidateBrand(v, brand); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Brands.Update(brand)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, brand, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteBrandHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Brands.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "brand successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listBrandsHandler(w http.ResponseWriter, r *http.Request) {
	brands, err := app.models.Brands.GetAll()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, brands, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
