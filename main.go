package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("GET /api/v1/records", handleGetRecord)
	http.HandleFunc("POST /api/v1/records", handlePostRecord)
	http.ListenAndServe(":8080", nil)
}

func handleGetRecord(w http.ResponseWriter, r *http.Request) {
	fmt.Println("im in")
	fmt.Fprintf(w, "print from insiders")
}

func handlePostRecord(w http.ResponseWriter, r *http.Request) {
	fmt.Println("im in")
	body := r.Body
	fmt.Fprintf(w, "print from insiders: %v", body)
}
