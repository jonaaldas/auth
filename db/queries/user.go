package queries

import (
	"database/sql"
	"errors"
	"fmt"
)

type User struct {
	ID             int    `json:"id"`
	Email          string `json:"email"`
	Username       string `json:"username"`
	HashedPassword string `json:"hashed_password"`
	Salt           string `json:"salt"`
}

type InsertUserData struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	HashedPassword string `json:"hashed_password"`
	Salt           string `json:"salt"`
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	row := db.QueryRow("SELECT id, email, username, hashed_password, salt FROM users WHERE email = ?", email)
	fmt.Println("row", row)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.HashedPassword, &user.Salt)

	if err != nil {
		return nil, err
	}

	if user.Email == "" {
		return nil, errors.New("user not found")
	}

	return &user, nil
}

func InsertUser(db *sql.DB, user InsertUserData) (bool, error) {
	_, err := db.Query("INSERT INTO users (username, email, hashed_password, salt) VALUES (?, ?, ?, ?)", user.Username, user.Email, user.HashedPassword, user.Salt)
	if err != nil {
		return false, err
	}
	return true, nil

}
