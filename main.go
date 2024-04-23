package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"eco-backend/endpoints"
)

type User struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar,omitempty"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("HOST"), os.Getenv("PORT"), os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("DBNAME"))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	http.Handle("/", corsMiddleware(http.DefaultServeMux))

	http.HandleFunc("/users/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		endpoints.GetUsers(w, r, db)
	})

	http.HandleFunc("/users/registration", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		endpoints.Registration(w, r, db)
	})

	http.HandleFunc("/users/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var authHeader string = r.Header.Get("Authorization")

		endpoints.Login(w, r, db, authHeader)
	})

	fmt.Println("Server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
