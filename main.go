package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	const rootDir = "."

	apiCfg := apiConfig{}

	serverMux := http.NewServeMux()

	// file server
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir(rootDir)))))

	serverMux.HandleFunc("GET /healthz", handlerReadiness)

	serverMux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	serverMux.HandleFunc("POST /admin/reset", apiCfg.handlerMetricsReset)

	serverMux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serverMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
