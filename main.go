package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jonaaldas/go-auth/auth"
	"github.com/jonaaldas/go-auth/db"
	userQueries "github.com/jonaaldas/go-auth/db/queries"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
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

	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: "c2VjcmV0LXRoaXJ0eS0yLWNoYXJhY3Rlci1zdHJpbmc=",
	}))

	app.Static("/", "./web/dist")

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

		existingUser, err := userQueries.GetUserByEmail(db, userLogin.EMAIL)

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

		token := auth.GenerateSessionToken()
		session, err := auth.CreateSession(db, token, existingUser.ID)
		if err != nil {
			return c.JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		c.Cookie(&fiber.Cookie{
			Name:     "session",
			Value:    session.Id,
			HTTPOnly: false,
			Secure:   false,
		})

		fmt.Print(session.Id)

		return c.JSON(fiber.Map{
			"success": true,
			"user":    existingUser,
		})
	})

	app.Post("/api/logout", func(c *fiber.Ctx) error {
		sessionID := c.Cookies("session")

		// Invalidate the session in the database if it exists
		if sessionID != "" {
			err := auth.InvalidateSession(db, sessionID)
			if err != nil {
				// Log the error but still clear the cookie
				fmt.Printf("Error invalidating session: %v\n", err)
			}
		}

		// Clear the session cookie
		c.Cookie(&fiber.Cookie{
			Name:     "session",
			Value:    "",
			Expires:  time.Now().Add(-time.Hour), // Set to past time to delete
			HTTPOnly: false,
			Secure:   false,
		})

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Logged out successfully",
		})
	})

	// Auth middleware function
	authMiddleware := func(c *fiber.Ctx) error {
		sessionID := c.Cookies("session")
		if sessionID == "" {
			return c.Status(401).JSON(fiber.Map{
				"error":   "Unauthorized - No session cookie",
				"success": false,
			})
		}

		// Get session from database and validate
		session, err := auth.ValidateSessionToken(db, sessionID)
		if err != nil || session.Id == "" {
			return c.Status(401).JSON(fiber.Map{
				"error":   "Unauthorized - Invalid session",
				"success": false,
				"message": err.Error(),
			})
		}

		// Get user from database
		user, err := userQueries.GetUserByID(db, session.UserId)
		if err != nil || user == nil {
			return c.Status(401).JSON(fiber.Map{
				"error":   "Unauthorized - User not found",
				"success": false,
			})
		}

		// Store user in locals for this request
		c.Locals("user", user)
		return c.Next()
	}

	// Protected routes - apply auth middleware only to these
	protected := app.Group("/api/protected")
	protected.Use(authMiddleware)

	protected.Get("/profile", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*userQueries.User)

		return c.JSON(fiber.Map{
			"success": true,
			"user":    user,
		})
	})

	log.Println("Server starting on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
