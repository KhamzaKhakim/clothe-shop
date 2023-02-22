package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/clothes", app.listClothesHandler)
	router.HandlerFunc(http.MethodPost, "/v1/clothes", app.requireRole("ADMIN", app.createClotheHandler))
	router.HandlerFunc(http.MethodGet, "/v1/clothes/:id", app.showClotheHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/clothes/:id", app.requireRole("ADMIN", app.updateClotheHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/clothes/:id", app.requireRole("ADMIN", app.deleteClotheHandler))

	router.HandlerFunc(http.MethodGet, "/v1/brands", app.listBrandsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/brands", app.requireRole("ADMIN", app.createBrandHandler))
	router.HandlerFunc(http.MethodGet, "/v1/brands/:id", app.showBrandHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/brands/:id", app.requireRole("ADMIN", app.updateBrandHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/brands/:id", app.requireRole("ADMIN", app.deleteBrandHandler))

	router.HandlerFunc(http.MethodPut, "/v1/buy/:id", app.requireRole("USER", app.addToCartHandler))
	router.HandlerFunc(http.MethodGet, "/v1/cart", app.requireRole("USER", app.showCartHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/users/:id", app.requireRole("ADMIN", app.deleteUserHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodGet, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
}
