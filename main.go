package main

import (
	"database/sql"
	"errors"
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

type FeedFollow struct {
	ID        uuid.UUID `json:"id"`
	FeedID    uuid.UUID `json:"feed_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	ApiKey    string    `json:"api_key"`
}

type Post struct {
	ID          uuid.UUID      `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Title       string         `json:"title"`
	Url         string         `json:"url"`
	Description sql.NullString `json:"description"`
	PublishedAt sql.NullTime   `json:"published_at"`
	FeedID      uuid.UUID      `json:"feed_id"`
}

var ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")

type Feed struct {
	ID            uuid.UUID  `json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Name          string     `json:"name"`
	Url           string     `json:"url"`
	UserID        uuid.UUID  `json:"user_id"`
	LastFetchedAt *time.Time `json:"last_fetched_at"`
}

func convertNullTimeToTimePtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func databaseFeedFollowsToFeedFollows(feedFollows []database.FeedFollow) []FeedFollow {
	result := make([]FeedFollow, len(feedFollows))
	for i, feedFollow := range feedFollows {
		result[i] = databaseFeedFollowToFeedFollow(feedFollow)
	}
	return result
}

func databaseFeedFollowToFeedFollow(feedFollow database.FeedFollow) FeedFollow {
	return FeedFollow{
		ID:        feedFollow.ID,
		FeedID:    feedFollow.FeedID,
		UserID:    feedFollow.UserID,
		CreatedAt: feedFollow.CreatedAt,
		UpdatedAt: feedFollow.UpdatedAt,
	}
}
func databaseFeedToFeed(feed database.Feed) Feed {
	return Feed{
		ID:            feed.ID,
		CreatedAt:     feed.CreatedAt,
		UpdatedAt:     feed.UpdatedAt,
		Name:          feed.Name,
		Url:           feed.Url,
		UserID:        feed.UserID,
		LastFetchedAt: convertNullTimeToTimePtr(feed.LastFetchedAt),
	}
}

func databaseUserToUser(user database.User) User {
	return User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		ApiKey:    user.ApiKey,
	}
}

func databasePostToPost(post database.Post) Post {

	return Post{
		ID:          post.ID,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		Title:       post.Title,
		Url:         post.Url,
		Description: post.Description,
		PublishedAt: post.PublishedAt,
		FeedID:      post.FeedID,
	}
}

func databasePostsToPosts(posts []database.Post) []Post {
	result := make([]Post, len(posts))
	for i, dbPost := range posts {
		result[i] = databasePostToPost(dbPost)
	}
	return result
}

func GetApiToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "ApiKey" {
		return "", errors.New("malformed authorization header")
	}

	return splitAuth[1], nil
}

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (cfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := GetApiToken(r.Header)
		if err != nil {
			respondWithERROR(w, http.StatusUnauthorized, "Couldn't find api key")
			return
		}

		user, err := cfg.DB.GetUserByApiKey(r.Context(), apiKey)
		if err != nil {
			respondWithERROR(w, http.StatusNotFound, "Couldn't get user")
			return
		}

		handler(w, r, user)
	}
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

	mux.HandleFunc("POST /v1/users", cfg.handlerPostUsers)
	mux.HandleFunc("GET /v1/users", cfg.middlewareAuth(cfg.handlerGetUsers))

	mux.HandleFunc("POST /v1/feeds", cfg.middlewareAuth(cfg.handlerPostFeeds))
	mux.HandleFunc("GET /v1/feeds", cfg.handlerGetFeeds)

	mux.HandleFunc("GET /v1/posts", cfg.middlewareAuth(cfg.handlerPostPost))

	mux.HandleFunc("POST /v1/feed_follows", cfg.middlewareAuth(cfg.handlerPostFeedFollows))
	mux.HandleFunc("DELETE /v1/feed_follows/{feedFollowID}", cfg.middlewareAuth(cfg.handlerDeleteFeedFollows))
	mux.HandleFunc("GET /v1/feed_follows", cfg.middlewareAuth(cfg.handlerFeedFollowsGet))

	mux.HandleFunc("GET /v1/healthz", handlerReadiness)
	mux.HandleFunc("GET /v1/err", handlerError)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	const collectionConcurrency = 10
	const collectionInterval = time.Minute
	go startScraping(dbQueries, collectionConcurrency, collectionInterval)

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
