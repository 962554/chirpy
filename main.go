// Package Chirpy is a social network similar to Twitter.
package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/962554/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const adminTemplate = `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`

var apiCfg = &apiConfig{platform: "dev"}

func main() {
	const (
		port        = ":8080"
		readTimeout = 5 * time.Second
	)
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error opening db: %s", err)
	}

	apiCfg.dbQueries = database.New(db)
	apiCfg.platform = "dev"

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadHeaderTimeout: readTimeout,
	}

	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateHandler)
	mux.HandleFunc("POST /api/users", createUsersHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.showHits)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHits)

	log.Printf("http server starting on port: %s", port)
	log.Fatal(server.ListenAndServe())
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
