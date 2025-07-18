package main

import (
	"log"
	"net/http"

	"go-local-hub/internal/clipboard"
	"go-local-hub/internal/metrics"
	"go-local-hub/internal/notes"
	"go-local-hub/internal/todos"
)

func main() {
	http.HandleFunc("/metrics", metrics.Handler)
	http.HandleFunc("/clipboard/set", clipboard.SetHandler)
	http.HandleFunc("/clipboard/get", clipboard.GetHandler)
	http.HandleFunc("/notes", notes.ListHandler)
	http.HandleFunc("/notes/save", notes.SaveHandler)
	http.HandleFunc("/todos", todos.ListHandler)
	http.HandleFunc("/todos/add", todos.AddHandler)
	http.Handle("/", http.FileServer(http.Dir("./web/dist")))

	log.Println("Listening on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
