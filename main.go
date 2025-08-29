package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type LoginRequest struct {
	Email string `json:"email"`
}

var db *sql.DB

func initDb() {
	url := os.Getenv("URL")
	fmt.Print(url)

	var err error
	db, err = sql.Open("libsql", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", url, err)
		os.Exit(1)
	}

	if err = db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to db: %s", err)
		os.Exit(1)
	}

	log.Println("Database connected successfully")
}

func main() {
	godotenv.Load()
	initDb()
	defer db.Close()

	app := fiber.New()

	app.Get("api/all", func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT * from users")
		if err != nil {
			log.Printf("Database query error: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
		defer rows.Close()

		return c.JSON(fiber.Map{
			"data": rows,
		})
	})

	app.Post("/api/login", func(c *fiber.Ctx) error {
		var loginReq LoginRequest

		rows, err := db.Query("SELECT id, email FROM users WHERE email = ?", loginReq.Email)
		if err != nil {
			log.Printf("Database query error: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.ID, &user.Email); err != nil {
				log.Printf("Row scan error: %v", err)
				continue
			}
			users = append(users, user)
		}

		if len(users) == 0 {
			return c.Status(404).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Login successful",
			"user":    users[0],
		})
	})

	log.Println("Server starting on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
