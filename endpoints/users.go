package endpoints

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type User struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar,omitempty"`
}

type Answer struct {
	IsOk    bool   `json:"isOk"`
	Message string `json:"message"`
}

func GetData(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
}

func InsertData(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var user User
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var result = Answer{IsOk: true, Message: "success"}

	var avatar *string
	if user.Avatar != nil {
		avatar = user.Avatar
	}
	_, err = db.Exec("INSERT INTO users (username, avatar) VALUES ($1, $2)", user.Username, avatar)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}
