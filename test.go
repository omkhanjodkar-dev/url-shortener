package main

import (
	"fmt"
	"net/http"
	"math/rand"
	"encoding/json"
)

func generateCode() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	code := make([]byte, 6)

	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}

	return string(code)
}

type Response struct {
	URL string `json:"url"`
}

func shorten(w http.ResponseWriter, r *http.Request, codes map[string]string) {
	if r.Method != http.MethodPost {
		http.Error(w, "wrong method", http.StatusMethodNotAllowed)
	}

	var req Response
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.URL == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	code := generateCode()
	_, exists := codes[code]
	for exists {
		code = generateCode()
		_, exists = codes[code]
	}

	codes[code] = req.URL

	fmt.Fprintf(w, code)
}

func code(w http.ResponseWriter, r *http.Request, codes map[string]string) {
	val, exists := codes[r.URL.Path[1:]]
	if exists {
		http.Redirect(w, r, val, http.StatusMovedPermanently)
		return
	}
	http.Error(w, "resource not found", http.StatusNotFound)
}

func main() {
	codes := map[string]string {}

	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request){shorten(w, r, codes)})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {code(w, r, codes)})

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err)
	}
}