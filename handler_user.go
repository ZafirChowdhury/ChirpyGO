package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Email string `json:"email"`
	}

	req := Request{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email cannot be empty")
		return
	}

	usr, err := cfg.db.CreateUser(r.Context(), req.Email)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, "Error while trying to create user")
		return
	}

	user := User{
		ID:        usr.ID,
		CreatedAt: usr.CreatedAt,
		UpdatedAt: usr.UpdatedAt,
		Email:     usr.Email,
	}

	respondWithJSON(w, http.StatusCreated, user)
}
