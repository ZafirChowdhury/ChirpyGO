package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ZafirChowdhury/ChirpyGO/internal/auth"
	"github.com/ZafirChowdhury/ChirpyGO/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	req := Request{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password is required")
		return
	}

	dbRes, err := cfg.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	match, err := auth.CheckPasswordHash(req.Password, dbRes.HashedPassword)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	if !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	user := User{
		ID:        dbRes.ID,
		CreatedAt: dbRes.CreatedAt,
		UpdatedAt: dbRes.UpdatedAt,
		Email:     dbRes.Email,
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	req := Request{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password is required")
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	usr, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: passwordHash,
	})
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
