package main

import (
	"encoding/json"
	"net/http"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type ValidateChirp struct {
		Body string `json:"body"`
	}

	type ReturnValue struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	vc := ValidateChirp{}
	err := decoder.Decode(&vc)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(vc.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	resJson := ReturnValue{
		Valid: true,
	}

	respondWithJSON(w, http.StatusOK, resJson)
}
