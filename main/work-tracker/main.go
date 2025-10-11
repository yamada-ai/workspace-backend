package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "ok"); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
		}
	})
	addr := ":8000"
	log.Printf("work-tracker listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
