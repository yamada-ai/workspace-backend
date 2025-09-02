package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})
	addr := ":8000"
	log.Printf("work-tracker listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
