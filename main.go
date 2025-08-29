package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

type Email struct {
	Email string `json:"email"`
}

func main() {
	app := fiber.New()
	app.Post("/api/login", func(c *fiber.Ctx) error {
		p := new(Email)
		if err := c.BodyParser(p); err != nil {
			return err
		}
		log.Println(p.Email) // jonaaldas@gmail.com
		return c.SendString("HELLO WORLD")
	})

	app.Listen(":3000")
}
