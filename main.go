package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/L-PDufour/Blog-aggr/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	ApiKey    string    `json:"api_key"`
}

var ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")

func databaseUserToUser(user database.User) User {
	return User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		ApiKey:    user.ApiKey,
	}
}

func GetApiToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	fmt.Println(authHeader)
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "ApiKey" {
		return "", errors.New("malformed authorization header")
	}

	return splitAuth[1], nil
}

func (cfg *apiConfig) handlerGetUsers(w http.ResponseWriter, r *http.Request) {
	apiKey, err := GetApiToken(r.Header)
	fmt.Println(apiKey)
	if err != nil {
		respondWithERROR(w, http.StatusUnauthorized, "Couldn't find api key")
		return
	}

	user, err := cfg.DB.GetUserByApiKey(r.Context(), apiKey)
	if err != nil {
		respondWithERROR(w, http.StatusNotFound, "Couldn't get user")
		return
	}

	respondWithJSON(w, http.StatusOK, databaseUserToUser(user))

}

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

func main() {
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	dbURL := os.Getenv("DB")
	if dbURL == "" {
		log.Fatalf("DB environment variable not set")
	}
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("PORT environment variable not set")
	}

	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)
	cfg := &apiConfig{
		DB: dbQueries,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/healthz", handlerReadiness)
	mux.HandleFunc("GET /v1/err", handlerError)
	mux.HandleFunc("POST /v1/users", cfg.handlerPostUsers)
	mux.HandleFunc("GET /v1/users", cfg.handlerGetUsers)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())

	// Print the PORT environment variable
}
