package endpoints

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar,omitempty"`
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

	var avatar *string
	if user.Avatar != nil {
		avatar = user.Avatar
	}
	_, err = db.Exec("INSERT INTO users (id, username, avatar) VALUES ($1, $2, $3)", user.ID, user.Username, avatar)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Данные успешно вставлены в базу данных")
}
