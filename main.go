package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"

	serverMux := http.NewServeMux()

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serverMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
