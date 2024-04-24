package endpoints

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	Password string  `json:"password"`
	Avatar   *string `json:"avatar,omitempty"`
}

type Answer struct {
	IsOk    bool   `json:"isOk"`
	Message string `json:"message"`
	Token   string `json:"token"`
}

var secretKey = []byte("secret")

func Login(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var user User
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var storedPassword string
	row := db.QueryRow("SELECT id, username, avatar, password FROM users WHERE username = $1", user.Username)

	err = row.Scan(&user.ID, &user.Username, &user.Avatar, &storedPassword)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password))
	if err != nil {
		http.Error(w, "Incorrect password", http.StatusUnauthorized)
		return
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["exp"] = time.Now().Add(time.Hour * 729).Unix()
	tokenString, err := token.SignedString(secretKey)
	
	if err != nil {
		http.Error(w, "Incorrect password", http.StatusUnauthorized)
		return
	}

	var result = Answer{IsOk: true, Message: "success", Token: tokenString}

	json.NewEncoder(w).Encode(result)
}

func Me(w http.ResponseWriter, r *http.Request, db *sql.DB, token string) {
	parsedToken, errJwt := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if errJwt != nil {
		return
	}

	if !parsedToken.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Failed to parse token claims", http.StatusInternalServerError)
		return
	}

	username, ok := claims["username"].(string)
	if !ok {
		http.Error(w, "Username not found in token claims", http.StatusInternalServerError)
		return
	}

	row := db.QueryRow("SELECT id, username, avatar FROM users WHERE username = $1", username)

	var user User

	err := row.Scan(&user.ID, &user.Username, &user.Avatar)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetUsers(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

func Registration(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var user User
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["exp"] = time.Now().Add(time.Hour * 729).Unix()
	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var result = Answer{IsOk: true, Message: "success", Token: tokenString}

	var avatar *string
	if user.Avatar != nil {
		avatar = user.Avatar
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (username, password, avatar) VALUES ($1, $2, $3)", user.Username, hashedPassword, avatar)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}
