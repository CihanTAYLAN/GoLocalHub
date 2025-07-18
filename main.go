package main

import (
	"log"
	"net/http"

	"go-local-hub/internal/metrics"
)

func main() {
	http.HandleFunc("/metrics", metrics.Handler)
	http.Handle("/", http.FileServer(http.Dir("./web/dist")))

	log.Println("Listening on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
