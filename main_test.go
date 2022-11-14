// main_test.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var ws WebService

func TestMain(m *testing.M) {
	ws = WebService{}
	ws.Init()

	ensureTableExists()

	code := m.Run()

	clearTable()

	os.Exit(code)
}

func ensureTableExists() {
	if _, err := ws.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	ws.DB.Exec("DELETE FROM cakes")
	ws.DB.Exec("ALTER TABLE cakes AUTO_INCREMENT = 1")
}

const tableCreationQuery = `
CREATE TABLE IF NOT EXISTS cakes (
	id INTEGER PRIMARY KEY AUTO_INCREMENT,
	title VARCHAR(100) NOT NULL,
    description VARCHAR(250),
    rating FLOAT,
    image VARCHAR(255),
    created_at DATETIME,
    updated_at DATETIME
)`

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/cakes", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	ws.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestGetNonExistentCake(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/cakes/45", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Cake not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Cake not found'. Got '%s'", m["error"])
	}
}

func TestCreateCake(t *testing.T) {
	clearTable()

	payload := []byte(`{"title":"titl","description":"desc","rating":4.4,"image":"url"}`)

	req, _ := http.NewRequest("POST", "/cakes", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["title"] != "titl" {
		t.Errorf("Expected title to be 'titl'. Got '%v'", m["title"])
	}

	if m["description"] != "desc" {
		t.Errorf("Expected description to be 'desc'. Got '%v'", m["description"])
	}

	if m["rating"] != 4.4 {
		t.Errorf("Expected rating to be '4.4'. Got '%v'", m["rating"])
	}

	if m["image"] != "url" {
		t.Errorf("Expected image to be 'url'. Got '%v'", m["image"])
	}

}

func addCake(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		statement := fmt.Sprintf("INSERT INTO cakes(title, description,rating,image,created_at,updated_at) VALUES('%s', '%s', %f, '%s', NOW())", ("Cake " + strconv.Itoa(i+1)), "desc", 4.4, "url")
		ws.DB.Exec(statement)
	}
}

func TestGetCake(t *testing.T) {
	clearTable()
	addCake(1)

	req, _ := http.NewRequest("GET", "/cakes/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateCake(t *testing.T) {
	clearTable()
	addCake(1)

	req, _ := http.NewRequest("GET", "/cakes/1", nil)
	response := executeRequest(req)
	var prevCake map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &prevCake)

	payload := []byte(`{"title":"titl new","description":"desc new","rating":6.6,"image":"url new"}`)

	req, _ = http.NewRequest("PUT", "/cakes/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != prevCake["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", prevCake["id"], m["id"])
	}

	if m["title"] == prevCake["title"] {
		t.Errorf("Expected the title to change from '%v' to '%v'. Got '%v'", prevCake["title"], m["title"], m["title"])
	}

}

func TestDeleteCake(t *testing.T) {
	clearTable()
	addCake(1)

	req, _ := http.NewRequest("GET", "/cakes/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/cakes/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/cakes/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}
