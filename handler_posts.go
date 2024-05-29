package main

import (
	"net/http"
	"strconv"

	"github.com/L-PDufour/Blog-aggr/internal/database"
)

func (cfg *apiConfig) handlerPostPost(w http.ResponseWriter, r *http.Request, user database.User) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if specifiedLimit, err := strconv.Atoi(limitStr); err == nil {
		limit = specifiedLimit
	}

	postList, err := cfg.DB.GetPostsForUser(r.Context(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		respondWithERROR(w, http.StatusInternalServerError, "Couldn't get feed follow")
		return
	}

	respondWithJSON(w, http.StatusOK, databasePostsToPosts(postList))

}
