package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/L-PDufour/Blog-aggr/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPostFeeds(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithERROR(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	feed, err := cfg.DB.CreateFeed(r.Context(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		Name:      params.Name,
		Url:       params.URL,
	})
	if err != nil {
		respondWithERROR(w, http.StatusInternalServerError, "Couldn't create feed")
		return
	}

	respondWithJSON(w, http.StatusOK, databaseFeedToFeed(feed))

}

func (cfg *apiConfig) handlerGetFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := cfg.DB.GetFeeds(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch feeds", http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, feeds)
}
