package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /shorten", func(w http.ResponseWriter, r* http.Request) {
		w.Write([]byte("Shorten URL endpoint TODO"))
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r* http.Request) {
		w.Write([]byte("Redirect endpoint TODO"))
	})

	fmt.Println("Server starting on :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}