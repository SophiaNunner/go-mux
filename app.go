package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // imported pq here because we need our application to work with PostgreSQL
)

// App holds our application
// exposes references to the router and the database that the application uses
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// init all routes for the implemented handlers (getProducts, createProduct etc)
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/products", a.getProducts).Methods("GET") // use the a.getProducts handler to handle GET requests at the /products endpoint
	a.Router.HandleFunc("/product", a.createProduct).Methods("POST")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.getProduct).Methods("GET")    // {id:[0-9]+}: Gorilla Mux should process a URL only if the id is a number
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.updateProduct).Methods("PUT") // and store the actual numeric value in the id variable
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.deleteProduct).Methods("DELETE")
}

// take in the details required to connect to the database.
// create a database connection and wire up the routes to respond according to the requirements.
// needed for running tests
func (a *App) Initialize(user, password, dbname string) {
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

// start the application
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(":8010", a.Router))
}

/**************** handler ****************/
// for each function (root) implemented in model.go (getProduct, createProduct etc)

// handler "getProduct" for the route in model.go that fetches a single product
// NOTE: include net/http and strconv modules
func (a *App) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"]) // retrieve the id of the product to be fetched from the requested URL
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p := product{ID: id}
	if err := p.getProduct(a.DB); err != nil { // call getProduct method from model.go to fetch the details of that product
		switch err { // some error occured when requesting -> give error message as response
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Product not found") // status code=404 (product not found)
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, p) // no error occured -> give json object as response
}

// handler "getProducts" for the route in model.go that fetches a list of products
// By default, start is set to 0 and count is set to 10.
// If these parameters aren’t provided, this handler will respond with the first 10 products.
func (a *App) getProducts(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	products, err := getProducts(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, products)
}

// handler "createProduct" for the route in model.go that creates/adds a product
// handler assumes that the request body is a JSON object containing the details of the product to be created.
// It extracts that object into a product and uses the createProduct method to create a product with these details.
func (a *App) createProduct(w http.ResponseWriter, r *http.Request) {
	var p product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := p.createProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}

// handler "updateProduct" for the route in model.go that updated a product and its details
// extract the product details from the request body. It also extracts the id from the URL
// and uses the id and the body to update the product in the database.
func (a *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var p product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updateProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

// handler "deleteProduct" for the route in model.go that deletes a product from the database
// extract the id from the requested URL and uses it to delete the corresponding product from the database.
func (a *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Product ID")
		return
	}

	p := product{ID: id}
	if err := p.deleteProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

/**************** response functions ****************/
// (needed for handler functions above)

// NOTE: include encoding/json module
// when some error occured while requesting
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message}) // create a json object with error message and code
}

// when no error occured while requesting
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
