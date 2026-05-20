package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	const rootDir = "."

	serverMux := http.NewServeMux()

	// file server
	serverMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(rootDir))))

	serverMux.HandleFunc("/healthz", handlerReadiness)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serverMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
