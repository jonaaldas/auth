package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type HashRes struct {
	HashedPassword []byte
	Salt           uint32
}

func generateRandomString() uint32 {
	return rand.Uint32()
}

type Session struct {
	Id        string
	UserId    int
	ExpiresAt time.Time
	CreatedAt time.Time
}

func HashPassword(password string) (HashRes, error) {
	randomString := generateRandomString()
	passwordSalt := fmt.Sprintf("%s%d", password, randomString)
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(passwordSalt),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return HashRes{}, err
	}

	return HashRes{
		HashedPassword: hashedPassword,
		Salt:           randomString,
	}, nil
}

func VerifyPassword(hashedPassword string, password string, salt string) bool {
	passwordSalt := fmt.Sprintf("%s%s", password, salt)
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(passwordSalt))

	if err != nil {
		return false
	}

	return true
}

func GenerateSessionToken() string {
	bytes := make([]byte, 20)
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}

func CreateSession(db *sql.DB, token string, userId int) (Session, error) {
	sessionId := sha256.Sum256([]byte(token))
	session := Session{
		Id:        base64.StdEncoding.EncodeToString(sessionId[:]),
		UserId:    userId,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
		CreatedAt: time.Now(),
	}
	_, err := db.Exec("INSERT INTO sessions (id, user_id, expires_at, created_at) VALUES (?, ?, ?, ?)", session.Id, session.UserId, session.ExpiresAt, time.Now())

	if err != nil {
		return Session{}, err
	}
	return session, nil
}

func ValidateSessionToken(db *sql.DB, token string) (Session, error) {
	row := db.QueryRow("SELECT sessions.id, sessions.user_id, sessions.expires_at, sessions.created_at FROM sessions WHERE id = ?", token)

	if row == nil {
		return Session{}, errors.New("session not found")
	}

	var session Session
	err := row.Scan(&session.Id, &session.UserId, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return Session{}, err
	}

	if time.Now().After(session.ExpiresAt) {
		return Session{}, errors.New("session expired")
	}

	if time.Now().Before(session.CreatedAt.Add(time.Hour * 24 * 15)) {
		_, err := db.Exec("UPDATE sessions SET expires_at = ? WHERE id = ?", time.Now().Add(time.Hour*24*30), session.Id)
		if err != nil {
			return Session{}, err
		}
		session.ExpiresAt = time.Now().Add(time.Hour * 24 * 30)
		return session, nil
	}

	return session, nil
}

func InvalidateSession(db *sql.DB, sessionId string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE id = ?", sessionId)
	if err != nil {
		return err
	}
	return nil
}

func InvalidateAllSessions(db *sql.DB, userId int) error {
	_, err := db.Exec("DELETE FROM sessions WHERE user_id = ?", userId)
	if err != nil {
		return err
	}
	return nil
}
