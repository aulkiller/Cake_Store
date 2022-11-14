package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	_ "github.com/mattes/migrate/source/file"
)

type WebService struct {
	DB     *sql.DB
	Router *mux.Router
}

func (ws *WebService) Init() {
	var err error
	ws.DB, err = sql.Open("mysql", "user:password@(127.0.0.1:3306)/cake_store_db?parseTime=true")
	if err != nil {
		panic("error: " + err.Error())
	}

	m, err := migrate.New(
		"file://db/migration",
		"mysql://user:password@(127.0.0.1:3306)/cake_store_db?parseTime=true")

	if err != nil {
		panic("error: " + err.Error())
	}

	m.Up()

	ws.Router = mux.NewRouter()
	ws.initRoutes()

	log.Println("Server running on port 8080")
	if err := http.ListenAndServe(":8080", ws.Router); err != nil {
		log.Fatalln("fatal error: ", err)
	}
}

func (ws *WebService) initRoutes() {
	ws.Router.HandleFunc("/cakes", ws.getCakes).Methods("GET")
	ws.Router.HandleFunc("/cakes", ws.createCake).Methods("POST")
	ws.Router.HandleFunc("/cakes/{id}", ws.getCake).Methods("GET")
	ws.Router.HandleFunc("/cakes/{id}", ws.updateCake).Methods("PATCH")
	ws.Router.HandleFunc("/cakes/{id}", ws.deleteCake).Methods("DELETE")
}

func (ws *WebService) getCakes(w http.ResponseWriter, r *http.Request) {
	cakes, err := getCakes(ws.DB)
	if err != nil {
		log.Println("error web: ", err)
		respondWithError(w, http.StatusAccepted, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, cakes)
}

func (ws *WebService) createCake(w http.ResponseWriter, r *http.Request) {
	var cake Cake
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&cake)
	if err != nil {
		log.Println("error web: ", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	err = cake.createCake(ws.DB)
	if err != nil {
		log.Println("error web: ", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cake.Created_at = cake.RawCreated_at.Format("2006-01-02 15:04:05")
	cake.Updated_at = cake.RawUpdated_at.Format("2006-01-02 15:04:05")

	respondWithJSON(w, http.StatusCreated, cake)
}

func (ws *WebService) getCake(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("error web: ", err)
		respondWithError(w, http.StatusBadRequest, "Invalid cake ID")
		return
	}

	cake := Cake{Id: id}
	if err := cake.getCake(ws.DB); err != nil {
		log.Println("error web: ", err)
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Cake not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	cake.Created_at = cake.RawCreated_at.Format("2006-01-02 15:04:05")
	cake.Updated_at = cake.RawUpdated_at.Format("2006-01-02 15:04:05")

	respondWithJSON(w, http.StatusOK, cake)
}

func (ws *WebService) updateCake(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("error web: ", err)
		respondWithError(w, http.StatusBadRequest, "Invalid cake ID")
		return
	}

	var cake Cake
	decoder := json.NewDecoder(r.Body)

	err = decoder.Decode(&cake)
	if err != nil {
		log.Println("error web: ", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	cake.Id = id

	err = cake.updateCake(ws.DB)
	if err != nil {
		log.Println("error web: ", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cake.Created_at = cake.RawCreated_at.Format("2006-01-02 15:04:05")
	cake.Updated_at = cake.RawUpdated_at.Format("2006-01-02 15:04:05")

	respondWithJSON(w, http.StatusOK, cake)
}

func (ws *WebService) deleteCake(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("error web: ", err)
		respondWithError(w, http.StatusBadRequest, "Invalid cake ID")
		return
	}

	cake := Cake{Id: id}
	if err := cake.deleteCake(ws.DB); err != nil {
		log.Println("error web: ", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "WebServicelication/json")
	w.WriteHeader(code)
	w.Write(response)
}
