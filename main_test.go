// Given that we will be running tests against a database, we need to ensure
// that the database is properly set up before any tests are run and is cleaned
// up after all tests have been finished

package main

import (
	"bytes"
	"encoding/json"
	"log"

	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var a App // references the main application (application we want to test)

// executed before all other tests to ensure cleanup of database etc
func TestMain(m *testing.M) {

	log.Println("This is a change in code for testing travisCI.")

	a.Initialize(
		os.Getenv("APP_DB_USERNAME"), // these must be set as environment variables
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	a.Initialize

	ensureTableExists() // before running tests, check availability of database table
	code := m.Run()     // run all tests
	clearTable()        // cleanup database
	os.Exit(code)
}

/**************** setup and housekeeping ****************/

// make sure that the table we need for testing is available
func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil { // tableCreationQuery is a constant in the database (defined below)
		log.Fatal(err) // log module must be imported
	}
}

// cleanup database
func clearTable() {
	a.DB.Exec("DELETE FROM products")
	a.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products
(
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)`

/**************** tests ****************/

// first test!
// test response to products endpoint with an empty table
// NOTE: add net/http module
// deletes all records from the products table and sends a GET request to the /products endpoint
func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req) // send HTTP request

	checkResponseCode(t, http.StatusOK, response.Code) // check if http response is what we expected

	if body := response.Body.String(); body != "[]" { // check that we receive an empty array (expected)
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

// try to fetch a product that does not exist
// NOTE: include encoding/json module
func TestGetNonExistentProduct(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/product/11", nil) // try to get product 11
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code) // check that status code=404 (product not found)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m) // extract json object from response
	if m["error"] != "Product not found" {    // check that error text is correctly "product not found"
		t.Errorf("Expected the 'error' key of the response to be set to 'Product not found'. Got '%s'", m["error"])
	}
}

// create a product
// manually add product to the database and then try to fetch it from the /product endpoint
// NOTE: include bytes module
func TestCreateProduct(t *testing.T) {
	clearTable()

	var jsonStr = []byte(`{"name":"test product", "price": 11.22}`)
	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr)) // post request to /product endpoint with the specified request content in jsonStr
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)                         // POST new product returns the created product as json object
	checkResponseCode(t, http.StatusCreated, response.Code) // check that status code=201 (resource created)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m) // extract json object from response

	// check that the returned json object is the actual object/product we've created
	if m["name"] != "test product" {
		t.Errorf("Expected product name to be 'test product'. Got '%v'", m["name"])
	}
	if m["price"] != 11.22 {
		t.Errorf("Expected product price to be '11.22'. Got '%v'", m["price"])
	}
	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
	}
}

// add a new product and then get/fetch it
func TestGetProduct(t *testing.T) {
	clearTable()
	addProducts(1) // add 1 product to the table

	req, _ := http.NewRequest("GET", "/product/1", nil) // fetch product 1
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code) // check status code=200 (success)
}

// add 1 or more products to the table for testing
// NOTE: include strconv module
func addProducts(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Product "+strconv.Itoa(i), (i+1.0)*10) // database query
	}
}

// add a product to the database and then update its detailed info
func TestUpdateProduct(t *testing.T) {
	clearTable()
	addProducts(1) // add 1 product

	// fetch product 1 from database
	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	var originalProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	// update database with prouct details
	var jsonStr = []byte(`{"name":"test product - updated name", "price": 11.22}`)
	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code) // check status code=200 (success)

	// check if responded json object contains the updated details
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["id"] != originalProduct["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], m["id"])
	}
	if m["name"] == originalProduct["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalProduct["name"], m["name"], m["name"])
	}
	if m["price"] == originalProduct["price"] {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalProduct["price"], m["price"], m["price"])
	}
}

// delete a product from the database
func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProducts(1) // add product 1

	req, _ := http.NewRequest("GET", "/product/1", nil) // try to fetch it
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code) // added+fetched successfully?

	req, _ = http.NewRequest("DELETE", "/product/1", nil) // delete product 1
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code) // request succeeded?

	req, _ = http.NewRequest("GET", "/product/1", nil) // try to fetch product 1 -> not found
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

/**************** helping methods for tests ****************/

// send HTTP  request
// use application's router and return the response
// NOTE: include net/httptest module
func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

// check if http response is what we expected
func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
