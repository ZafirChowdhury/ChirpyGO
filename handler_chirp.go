package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ZafirChowdhury/ChirpyGO/internal/database"
	"github.com/google/uuid"
)

func dbResToChirp(dbRes database.Chirp) Chirp {
	return Chirp{
		ID:        dbRes.ID,
		CreatedAt: dbRes.CreatedAt,
		UpdatedAt: dbRes.UpdatedAt,
		Body:      dbRes.Body,
		UserID:    dbRes.UserID,
	}

}

func (cfg *apiConfig) handlerCreateNewChirp(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		UserID uuid.UUID `json:"user_id"`
		Body   string    `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	res := Response{}
	err := decoder.Decode(&res)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// check length
	if len(res.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	dbRes, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		UserID: res.UserID,
		Body:   profanityFilter(res.Body),
	})
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, "There was a error while tryring to save the chirp")
		return
	}

	c := dbResToChirp(dbRes)
	respondWithJSON(w, http.StatusCreated, c)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	dbRes, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, "There was a error while trying to get chrips")
		return
	}

	chirps := []Chirp{}
	for _, c := range dbRes {
		chirps = append(chirps, dbResToChirp(c))
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusNotFound, "Not found")
		return
	}

	dbRes, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusNotFound, "Not found")
		return
	}

	c := dbResToChirp(dbRes)
	respondWithJSON(w, http.StatusOK, c)
}
