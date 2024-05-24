package main

import "net/http"

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	type bodyResponse struct {
		Status string `json:"status"`
	}
	respondWithJSON(w, 200, bodyResponse{
		Status: "ok",
	})
}

func handlerError(w http.ResponseWriter, r *http.Request) {
	respondWithERROR(w, 500, "Internal Server Error")
}
