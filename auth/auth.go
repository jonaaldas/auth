package auth

import (
	"fmt"
	"math/rand"

	"golang.org/x/crypto/bcrypt"
)

type HashRes struct {
	HashedPassword []byte
	Salt           uint32
}

func generateRandomString() uint32 {
	return rand.Uint32()
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
