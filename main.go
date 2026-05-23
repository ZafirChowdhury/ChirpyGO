package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/ZafirChowdhury/ChirpyGO/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func main() {
	const port = "8080"
	const rootDir = "."

	// database
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       os.Getenv("PLATFORM"),
	}

	serverMux := http.NewServeMux()

	// file server
	serverMux.Handle("/app/", cfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir(rootDir)))))

	serverMux.HandleFunc("GET /healthz", handlerReadiness)

	serverMux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	serverMux.HandleFunc("POST /admin/reset", cfg.handlerAdminReset)

	serverMux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	serverMux.HandleFunc("POST /api/users", cfg.handlerCreateUser)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serverMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
