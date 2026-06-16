package main

import (
	"math/rand"
	"encoding/json"
	"fmt"
	"net/http"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

func generateCode() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	code := make([]byte, 6)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}

	return string(code)
}

func shorten(w http.ResponseWriter, r *http.Request, codes map[string]string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.URL == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	code := generateCode()
	_, exists := codes[code]
	for exists {
		code = generateCode()
		_, exists = codes[code]
	}

	codes[code] = req.URL

	fmt.Fprintf(w, "%s", "localhost:8080/" + string(code))
}

func def(w http.ResponseWriter, r *http.Request, codes map[string]string) {
	val, exists := codes[r.URL.Path[1:]]
	if exists {
		http.Redirect(w, r, val, http.StatusMovedPermanently)
		return
	}
	http.Error(w, "Resource not found", http.StatusNotFound)
}

func main() {
	codes := map[string]string{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {def(w, r, codes)})
	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {shorten(w, r, codes)})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}