package main

import (
	"bytes"
	"clothing-store/internal/data"
	"clothing-store/internal/jsonlog"
	"clothing-store/internal/mailer"
	"clothing-store/internal/validator"
	"context"
	_ "database/sql"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	_ "log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

var testApp application

func init() {
	testApp = GetApp()
}

func GetApp() application {

	var cfg config
	cfg.port = 4001
	cfg.env = "development"
	cfg.db.dsn = "postgres://postgres:1234@localhost/clothe_shop_test?sslmode=disable"

	cfg.db.maxOpenConns = 25
	cfg.db.maxIdleConns = 25
	cfg.db.maxIdleTime = "15m"

	cfg.smtp.host = "sandbox.smtp.mailtrap.io"
	cfg.smtp.port = 2525
	cfg.smtp.username = "ff000c31a6a652"
	cfg.smtp.password = "ec70d281de0e41"
	cfg.smtp.sender = "noreply@clotheshop.com"

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	//defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	//err = app.serve()
	//if err != nil {
	//	logger.PrintFatal(err, nil)
	//}

	return *app
}

func TestValidateUser(t *testing.T) {

	user := &data.User{
		Name:      "Khamza",
		Email:     "Khamza",
		Activated: false,
		Money:     100000,
	}

	user.Password.Set("test1111!1")

	v := validator.New()
	if data.ValidateUser(v, user); v.Valid() {
		t.Errorf("%v", v.Errors)
	}

	user1 := &data.User{
		Name:      "Khamza",
		Email:     "user@gmail.com",
		Activated: false,
		Money:     100000,
	}
	user1.Password.Set("test22222222222!")

	v1 := validator.New()
	if data.ValidateUser(v1, user1); !v1.Valid() {
		t.Errorf("%v", v.Errors)
	}

}

func TestValidateBrand(t *testing.T) {

	correctBrand := &data.Brand{
		Name:        "Adidas",
		Country:     "Germany",
		Description: "Blablabla",
		ImageURL:    "img.jpeg",
	}

	v := validator.New()
	data.ValidateBrand(v, correctBrand)

	if !v.Valid() {
		t.Errorf("%v", v.Errors)
	}

	incorrectBrand := &data.Brand{
		Name:     "Adidas",
		Country:  "Germany",
		ImageURL: "img.jpeg",
	}

	v1 := validator.New()

	data.ValidateBrand(v1, incorrectBrand)

	if v1.Valid() {
		t.Errorf("%v", v1.Errors)
	}

}

func TestValidatePermittedValue(t *testing.T) {
	value := "s"
	value1 := "small"

	permittedValues := []string{"s", "m", "l"}

	v := validator.New()
	v.Check(validator.PermittedValue(value, permittedValues...), "size", "invalid size")
	if !v.Valid() {
		t.Errorf("%v", v.Errors)
	}

	v1 := validator.New()
	v1.Check(validator.PermittedValue(value1, permittedValues...), "size", "invalid size")
	if v1.Valid() {
		t.Errorf("Have to get error, unexpected result")
	}
}

func TestReadJSON(t *testing.T) {
	//app := GetApp()
	var input struct {
		Name string `json:"name"`
	}

	jsonBody := []byte(`{"name": "Khamza"}`)
	bodyReader := bytes.NewReader(jsonBody)
	w := new(http.ResponseWriter)
	r, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/v1/clothes/4", bodyReader)
	err := testApp.readJSON(*w, r, &input)
	if err != nil {
		t.Errorf("Got an unexpected error: %v", err)
	}
}

func TestReadCSV(t *testing.T) {
	//app := GetApp()
	url1, _ := url.Parse("http://localhost:8080?sizes=xs,s,m")
	urlValues := url1.Query()
	sizes := testApp.readCSV(urlValues, "sizes", []string{})

	if len(sizes) != 3 {
		t.Errorf("Expected array with size 3, but got %d", len(sizes))
	}
	for i := 0; i < len(sizes); i++ {
		if sizes[i] != "xs" && sizes[i] != "s" && sizes[i] != "m" {
			t.Errorf("Unexpected value %v", sizes[i])
		}
	}
}

func TestReadString(t *testing.T) {
	//app := GetApp()
	url1, err := url.Parse("http://localhost:8080?key=adidas")
	urlValues := url1.Query()
	key := testApp.readString(urlValues, "key", "")
	if key != "adidas" || err != nil {
		t.Errorf("Expected value is adidas, but got %v", key)
	}
}

func TestReadInt(t *testing.T) {
	//app := GetApp()
	r, _ := url.Parse("http://localhost:8080?price=50000")
	v := validator.New()
	qs := r.Query()
	price := testApp.readInt(qs, "price", 10000000, v)

	if price != 50000 {
		t.Errorf("Expected value is 50000, but got %d", price)
	}
}

func TestReadID(t *testing.T) {
	//app := GetApp()
	req := httptest.NewRequest(http.MethodGet, "/v1/clothes", nil)

	params := httprouter.Params{
		{Key: "id", Value: "1"},
	}
	ctx := context.WithValue(req.Context(), httprouter.ParamsKey, params)
	req = req.WithContext(ctx)
	id, err := testApp.readIDParam(req)

	if id != 1 || err != nil {
		t.Errorf("Expected 1, got %v", id)
	}
}

func TestGetAllClothes(t *testing.T) {
	//app := GetApp()
	req, err := http.NewRequest("GET", "/v1/clothes", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testApp.listClothesHandler)

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var arr []data.Clothe
	var arr1 []data.Clothe

	response := rr.Body.String()

	err = json.Unmarshal([]byte(response), &arr)
	if err != nil {
		t.Errorf("Can't marshall response to clothe type")
	}

	clothe := &data.Clothe{
		Name:     "test",
		Price:    1,
		Brand:    "test",
		Color:    "test",
		Sizes:    []string{},
		Sex:      "",
		Type:     "",
		ImageURL: "",
	}

	testApp.models.Clothes.Insert(clothe)

	rr1 := httptest.NewRecorder()

	handler.ServeHTTP(rr1, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	secondResponse := rr1.Body.String()
	err = json.Unmarshal([]byte(secondResponse), &arr1)
	if err != nil {
		t.Errorf("Can't marshall response to clothe type")
	}

	if len(arr1)-len(arr) != 1 {
		t.Errorf("Data length before insertion %v, after %v", len(arr), len(arr1))
	}
}

func TestAuthorization(t *testing.T) {
	//app := GetApp()

	req := httptest.NewRequest(http.MethodGet, "/v1/clothes", nil)
	req.Header.Set("Authorization", "Bearer teststesse")

	w := httptest.NewRecorder()

	handler := testApp.authenticate(http.HandlerFunc(testApp.listClothesHandler))
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected to get status code 401, but got %v", w.Code)
	}

}

func TestRequireRole(t *testing.T) {
	//app := GetApp()

	req := httptest.NewRequest(http.MethodGet, "/v1/clothes", nil)
	req.Header.Set("Authorization", "Bearer teststesse")

	user := &data.User{
		ID:        1,
		Activated: true,
	}

	req = testApp.contextSetUser(req, user)
	w := httptest.NewRecorder()

	handler := testApp.requireRole("ADMIN", testApp.listBrandsHandler)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected to get status code 200, but got %v", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/clothes", nil)
	req.Header.Set("Authorization", "Bearer teststesse")

	user = &data.User{
		ID:        2,
		Activated: true,
	}

	req = testApp.contextSetUser(req, user)

	w1 := httptest.NewRecorder()

	handler = testApp.requireRole("ADMIN", testApp.listBrandsHandler)
	handler.ServeHTTP(w1, req)

	if w1.Code != http.StatusForbidden {
		t.Errorf("Expected to get status code 200, but got %v", w1.Code)
	}
}
