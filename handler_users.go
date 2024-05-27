package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/L-PDufour/Blog-aggr/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPostUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name string
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithERROR(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      params.Name,
	})

	if err != nil {
		respondWithERROR(w, http.StatusInternalServerError, "Couldn't create user")
	}
	respondWithJSON(w, http.StatusCreated, databaseUserToUser(user))

}

func (cfg *apiConfig) handlerGetUsers(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, http.StatusOK, databaseUserToUser(user))
}
