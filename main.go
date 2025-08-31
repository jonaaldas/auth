package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/jonaaldas/go-auth/auth"
	"github.com/jonaaldas/go-auth/db"
	userQueries "github.com/jonaaldas/go-auth/db/queries"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type LoginRequest struct {
	EMAIL    string `json:"EMAIL"`
	PASSWORD string `json:"PASSWORD"`
}

type RegistrationBody struct {
	NAME     string `json:"name"`
	EMAIL    string `json:"email"`
	PASSWORD string `json:"password"`
}

func main() {
	godotenv.Load()
	db := db.InitDb()
	defer db.Close()

	app := fiber.New()

	app.Post("api/register", func(c *fiber.Ctx) error {
		user := new(RegistrationBody)
		if err := c.BodyParser(user); err != nil {
			return err
		}

		existingUser, err := userQueries.GetUserByEmail(db, user.EMAIL)

		if err == nil && existingUser != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "User already exists",
			})
		}

		res, err := auth.HashPassword(user.PASSWORD)

		if err != nil {
			log.Fatalf("Error hashing password: %v", err)
		}

		_, err = userQueries.InsertUser(db, userQueries.InsertUserData{
			Username:       user.NAME,
			Email:          user.EMAIL,
			HashedPassword: string(res.HashedPassword),
			Salt:           strconv.Itoa(int(res.Salt)),
		})

		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
		})
	})

	app.Post("/api/login", func(c *fiber.Ctx) error {
		userLogin := new(LoginRequest)
		if err := c.BodyParser(userLogin); err != nil {
			return err
		}
		fmt.Println("userLogin", userLogin)
		existingUser, err := userQueries.GetUserByEmail(db, userLogin.EMAIL)
		fmt.Println(existingUser)

		if err != nil || existingUser == nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid email or password",
			})
		}

		verify := auth.VerifyPassword(existingUser.HashedPassword, userLogin.PASSWORD, existingUser.Salt)
		if !verify {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid email or password",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
		})
	})

	log.Println("Server starting on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
