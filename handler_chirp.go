package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ZafirChowdhury/ChirpyGO/internal/database"
	"github.com/google/uuid"
)

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

	c := Chirp{
		ID:        dbRes.ID,
		CreatedAt: dbRes.CreatedAt,
		UpdatedAt: dbRes.UpdatedAt,
		Body:      dbRes.Body,
		UserID:    dbRes.UserID,
	}

	respondWithJSON(w, http.StatusCreated, c)
}
