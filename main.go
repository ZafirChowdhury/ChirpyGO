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
	secretKey      string
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
		secretKey:      os.Getenv("SECRET_KEY"),
	}

	serverMux := http.NewServeMux()

	// file server
	serverMux.Handle("/app/", cfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir(rootDir)))))

	// server health
	serverMux.HandleFunc("GET /healthz", handlerReadiness)

	// admin
	serverMux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	serverMux.HandleFunc("POST /admin/reset", cfg.handlerAdminReset)

	// users
	serverMux.HandleFunc("POST /api/login", cfg.handlerLogin)
	serverMux.HandleFunc("POST /api/users", cfg.handlerCreateUser)
	serverMux.HandleFunc("PUT /api/users", cfg.handlerUpdateUser)

	// auth
	serverMux.HandleFunc("POST /api/refresh", cfg.handlerRefresh)
	serverMux.HandleFunc("POST /api/revoke", cfg.handlerRevoke)

	// chirps
	serverMux.HandleFunc("GET /api/chirps", cfg.handlerGetChirps)
	serverMux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerGetChirp)
	serverMux.HandleFunc("POST /api/chirps", cfg.handlerCreateNewChirp)
	serverMux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.handlerDeleteChirp)

	// webhook
	serverMux.HandleFunc("POST /api/polka/webhooks", cfg.handlerUpgradeUserToRed)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serverMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
