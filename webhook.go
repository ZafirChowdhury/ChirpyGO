package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ZafirChowdhury/ChirpyGO/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUpgradeUserToRed(w http.ResponseWriter, r *http.Request) {
	api, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Println("Invalid Polka API key!")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if api != cfg.polkaAPI {
		log.Println("Invalid Polka API key!")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	type JBody struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	jBody := JBody{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&jBody); err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if jBody.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	uID, err := uuid.Parse(jBody.Data.UserID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = cfg.db.UpgradeUserToRed(r.Context(), uID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
