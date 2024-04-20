package main

import (
    "database/sql"
	"encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/joho/godotenv"
    _ "github.com/lib/pq"
)

type User struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
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

	http.HandleFunc("/getdata", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, username, avatar FROM users")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var user User
			var avatar sql.NullString
			err := rows.Scan(&user.ID, &user.Username, &avatar)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if avatar.Valid {
				user.Avatar = &avatar.String 
			}
			users = append(users, user)
		}
		json.NewEncoder(w).Encode(users)
	})

	http.HandleFunc("/insertdata", func(w http.ResponseWriter, r *http.Request) {
        var user User
        decoder := json.NewDecoder(r.Body)
        err := decoder.Decode(&user)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        var avatar *string
        if user.Avatar != nil {
            avatar = &(*user.Avatar)
        }
        _, err = db.Exec("INSERT INTO users (username, avatar) VALUES ($1, $2)", user.Username, avatar)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        fmt.Fprintf(w, "Данные успешно вставлены в базу данных")
    })

    fmt.Println("Server listening on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}