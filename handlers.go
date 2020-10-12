package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/ssrdive/basara/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// user := app.extractUser(r)

	if app.runtimeEnv == "dev" {
		fmt.Fprintf(w, "It works! [dev]")
	} else {
		fmt.Fprintf(w, "It works!")
	}
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	u, err := app.user.Get(username, password)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["username"] = u.Username
	claims["name"] = u.Name
	claims["exp"] = time.Now().Add(time.Minute * 180).Unix()

	ts, err := token.SignedString(app.secret)
	if err != nil {
		app.serverError(w, err)
		return
	}

	user := models.UserResponse{u.ID, u.Username, u.Name, u.Type, ts}
	js, err := json.Marshal(user)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (app *application) dropdownHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	if name == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	items, err := app.dropdown.Get(name)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)

}

func (app *application) dropdownConditionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	where := vars["where"]
	value := vars["value"]
	if name == "" || where == "" || value == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	items, err := app.dropdown.ConditionGet(name, where, value)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)

}

func (app *application) itemTest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Item Test")
}

func (app *application) itemSearch(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	results, err := app.item.Search(search)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (app *application) itemDetailsById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	items, err := app.item.DetailsById(id)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)

}

func (app *application) itemDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	items, err := app.item.Details(id)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)

}

func (app *application) updateItemById(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"item_id", "item_name", "item_price"}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			fmt.Println(param)
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	id, err := app.item.UpdateById(r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%d", id)
}

func (app *application) createBusinessPartner(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"user_id", "business_partner_type_id", "name", "address", "telephone", "email"}
	optionalParams := []string{}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			fmt.Println(param)
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	id, err := app.businessPartner.Create(requiredParams, optionalParams, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%d", id)

}

func (app *application) createItem(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	requiredParams := []string{"user_id", "item_id", "model_id", "item_category_id", "page_no", "item_no", "foreign_id", "item_name", "price"}
	optionalParams := []string{}
	for _, param := range requiredParams {
		if v := r.PostForm.Get(param); v == "" {
			fmt.Println(param)
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	id, err := app.item.Create(requiredParams, optionalParams, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%d", id)

}

func (app *application) allItems(w http.ResponseWriter, r *http.Request) {
	results, err := app.item.All()
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
