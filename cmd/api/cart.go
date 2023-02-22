package main

import (
	"clothing-store/internal/data"
	"clothing-store/internal/validator"
	"errors"
	"net/http"
	"strings"
)

func (app *application) addToCartHandler(w http.ResponseWriter, r *http.Request) {
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

	authorizationHeader := r.Header.Get("Authorization")

	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}

	token := headerParts[1]
	v := validator.New()
	if data.ValidateTokenPlaintext(v, token); !v.Valid() {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidAuthenticationTokenResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.models.Users.UpdateMoney(user, clothe.Price)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Carts.AddClotheForCart(user.ID, *clothe)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "clothe successfully added to the cart"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showCartHandler(w http.ResponseWriter, r *http.Request) {

	var response struct {
		Name      string  `json:"name"`
		Money     int64   `json:"money"`
		ClothesID []int64 `json:"clothes_id"`
	}

	authorizationHeader := r.Header.Get("Authorization")

	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}

	token := headerParts[1]
	v := validator.New()
	if data.ValidateTokenPlaintext(v, token); !v.Valid() {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidAuthenticationTokenResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	clothes, err := app.models.Carts.GetById(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	response.Name = user.Name
	response.Money = user.Money
	response.ClothesID = clothes

	err = app.writeJSON(w, http.StatusOK, response, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
