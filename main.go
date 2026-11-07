package main

import (
	"os"
	"sync"
	"net"
	"strings"
	"time"
	"strconv"
	"math/rand"
	"encoding/json"
	"fmt"
	"net/http"

	"database/sql"
	_ "modernc.org/sqlite"
)

type RateLimiter struct {
	mu sync.Mutex
	rs map[string]Bucket
}

type Bucket struct {
	LastReqTime time.Time
	Tokens int
}

type ShortenRequest struct {
	URL string `json:"url"`
	CustomCode string `json:"code"`
}

type Entry struct {
	code string
	URL string
}

func read(db *sql.DB, code string) (string, error) {
	var url string
	err := db.QueryRow("SELECT url FROM codes WHERE code = ?", code).Scan(&url)
	return url, err
}

func readExpiry(db *sql.DB, code string) (int, error) {
	var expiry string
	err := db.QueryRow("SELECT expiry FROM codes WHERE code = ?", code).Scan(&expiry)
	expiryInt, err := strconv.Atoi(expiry)
	return expiryInt, err
}

func write(db *sql.DB, code string, url string, expiry int64) error {
	_, err := db.Exec("INSERT INTO codes (code, url, expiry) VALUES (?, ?, ?)", code, url, expiry)
	return err
}

func IncClicks(db *sql.DB, code string) error {
	_, err := db.Exec("UPDATE codes SET clicks = clicks + 1 WHERE code = ?", code)
	return err
}

func GetClicks(db *sql.DB, code string) (int, error) {
	var clicks int
	err := db.QueryRow("SELECT clicks FROM codes WHERE code = ?", code).Scan(&clicks)
	return clicks, err
}

func generateCode() string{
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	code := make([]byte, 6)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}

	return string(code)
}

func getIP(r *http.Request) string {
    // Check proxy headers first
    if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
        return strings.Split(ip, ",")[0] // may contain multiple IPs
    }
    if ip := r.Header.Get("X-Real-Ip"); ip != "" {
        return ip
    }
    // Fall back to RemoteAddr (strips the port)
    ip, _, _ := net.SplitHostPort(r.RemoteAddr)
    return ip
}

func shorten(w http.ResponseWriter, r *http.Request, db *sql.DB, limiter *RateLimiter) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}


	limiter.mu.Lock()

	val, exists := limiter.rs[getIP(r)]
	fmt.Println(val)
	if exists {
		elapsed := time.Since(val.LastReqTime)
		fmt.Println(elapsed)
		val.LastReqTime = time.Now()
		if val.Tokens < 10 {
			val.Tokens = val.Tokens + int(elapsed.Seconds() * 1)
		}
		if val.Tokens >= 1 {
			val.Tokens = val.Tokens - 1
		} else {
			limiter.mu.Unlock()
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		limiter.rs[getIP(r)] = val
	} else {
		var new_buck Bucket
		new_buck.Tokens = 9
		new_buck.LastReqTime = time.Now()
		limiter.rs[getIP(r)] = new_buck
	}

	limiter.mu.Unlock()


	var req ShortenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.URL == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.CustomCode == "" {
		code := generateCode()
		_, exists := read(db, code)
		for exists == nil {
			code = generateCode()
			_, exists = read(db, code)
		}
	
		err = write(db, code, req.URL, time.Now().Add(time.Minute).Unix())
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		
		fmt.Fprintf(w, "%s", "localhost:8080/" + string(code))
		return
	} else {
		_, err = read(db, req.CustomCode)
		if err == nil {
			http.Error(w, "code already exists", http.StatusBadRequest)
			return
		} else {
			err = write(db, req.CustomCode, req.URL, time.Now().Add(time.Minute).Unix())
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}

			fmt.Fprintf(w, "%s", "localhost:8080/" + string(req.CustomCode))
			return
		}
	}

}

func def(w http.ResponseWriter, r *http.Request, db *sql.DB, limiter *RateLimiter) {
	

	limiter.mu.Lock()

	val, exists := limiter.rs[getIP(r)]
	fmt.Println(val)
	if exists {
		elapsed := time.Since(val.LastReqTime)
		fmt.Println(elapsed)
		val.LastReqTime = time.Now()
		if val.Tokens < 10 {
			val.Tokens = val.Tokens + int(elapsed.Seconds() * 1)
		}
		if val.Tokens >= 1 {
			val.Tokens = val.Tokens - 1
		} else {
			limiter.mu.Unlock()
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		limiter.rs[getIP(r)] = val
	} else {
		var new_buck Bucket
		new_buck.Tokens = 9
		new_buck.LastReqTime = time.Now()
		limiter.rs[getIP(r)] = new_buck
	}

	limiter.mu.Unlock()



	
	val2, exists2 := read(db, r.URL.Path[1:])
	if exists2 == nil {
		

		expiry, err := readExpiry(db, r.URL.Path[1:])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if time.Now().Unix() > int64(expiry) {
			http.Error(w, "resource expired", http.StatusGone)
			return
		}


		err = IncClicks(db, r.URL.Path[1:])
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			fmt.Println(err)
			return
		}
		fmt.Println(r.URL.Path[1:])
		http.Redirect(w, r, val2, http.StatusFound)
		return
	}
	http.Error(w, "Resource not found", http.StatusNotFound)
}

type CodeJSON struct {
	Code string `json:"code"`
}

func CheckClicks(w http.ResponseWriter, r *http.Request, db *sql.DB, limiter *RateLimiter) {	
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method type", http.StatusMethodNotAllowed)
		return
	}

	
	limiter.mu.Lock()

	val, exists := limiter.rs[getIP(r)]
	fmt.Println(val)
	if exists {
		elapsed := time.Since(val.LastReqTime)
		fmt.Println(elapsed)
		val.LastReqTime = time.Now()
		if val.Tokens < 10 {
			val.Tokens = val.Tokens + int(elapsed.Seconds() * 1)
		}
		if val.Tokens >= 1 {
			val.Tokens = val.Tokens - 1
		} else {
			limiter.mu.Unlock()
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		limiter.rs[getIP(r)] = val
	} else {
		var new_buck Bucket
		new_buck.Tokens = 9
		new_buck.LastReqTime = time.Now()
		limiter.rs[getIP(r)] = new_buck
	}

	limiter.mu.Unlock()


	var req CodeJSON
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Code == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	val2, err := GetClicks(db, req.Code)
	if err != nil {
		http.Error(w, "resource not found", http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	fmt.Fprintf(w, strconv.Itoa(val2))
}

func initDB(db *sql.DB) {
	query := "CREATE TABLE IF NOT EXISTS codes (code TEXT PRIMARY KEY, url TEXT NOT NULL, clicks INTEGER DEFAULT 0, expiry INTEGER)"
	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func main() {
	// rs := map[string]Bucket{}
	limiter := &RateLimiter{
		rs: make(map[string]Bucket),
	}

	db, err := sql.Open("sqlite", "codes.db")
	if err != nil {
		panic(err)
	}
	
	defer db.Close()

	_, err = db.Exec("DROP TABLE codes")

	initDB(db)

	// codes := map[string]string{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {def(w, r, db, limiter)})
	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {shorten(w, r, db, limiter)})
	http.HandleFunc("/clicks", func(w http.ResponseWriter, r *http.Request) {CheckClicks(w, r, db, limiter)})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	err = http.ListenAndServe(":" + port, nil)
	if err != nil {
		panic(err)
	}
}