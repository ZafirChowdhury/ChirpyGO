package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ZafirChowdhury/ChirpyGO/internal/auth"
	"github.com/ZafirChowdhury/ChirpyGO/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Password     string `json:"password"`
		Email        string `json:"email"`
		ExpiresInSec int    `json:"expires_in_seconds"`
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

	expiresIn := time.Duration(req.ExpiresInSec) * time.Second
	if expiresIn <= 0 || expiresIn > time.Hour {
		expiresIn = time.Hour
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

	token, err := auth.MakeJWT(dbRes.ID, cfg.secretKey, expiresIn)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	refreshToken := auth.MakeRefreshToken()
	rt, err := cfg.db.SaveRefreshToken(r.Context(), database.SaveRefreshTokenParams{
		Token:     refreshToken,
		UserID:    dbRes.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	user := User{
		ID:           dbRes.ID,
		CreatedAt:    dbRes.CreatedAt,
		UpdatedAt:    dbRes.UpdatedAt,
		Email:        dbRes.Email,
		Token:        token,
		RefreshToken: rt.Token,
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

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	rt, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		// dosent exist
		log.Println(err.Error())
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if rt.RevokedAt.Valid || rt.ExpiresAt.Before(time.Now().UTC()) {
		// revoked or expired
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	type Return struct {
		Token string `json:"token"`
	}

	jwt, err := auth.MakeJWT(rt.UserID, cfg.secretKey, time.Hour)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	returnBody := Return{
		Token: jwt,
	}

	respondWithJSON(w, http.StatusOK, returnBody)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = cfg.db.RevokeToken(r.Context(), token)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type JBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	jwtToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	uID, err := auth.ValidateJWT(jwtToken, cfg.secretKey)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	jBody := JBody{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&jBody); err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusUnauthorized, "Invalid JSON body")
		return
	}

	_, err = cfg.db.GetUserByID(r.Context(), uID)
	if err != nil {
		log.Println("User not found in database")
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	newPasswordHash, err := auth.HashPassword(jBody.Password)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	u, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          jBody.Email,
		HashedPassword: newPasswordHash,
		ID:             uID,
	})
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	user := User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
	}

	respondWithJSON(w, http.StatusOK, user)
}
