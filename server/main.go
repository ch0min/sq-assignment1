package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/oTuff/sq-ola1/db"
	"github.com/oTuff/sq-ola1/routes"
	"log"
)

func main() {
	// Setup database
	database, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Setup Fiber app
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Attach the database to the context
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", database)
		return c.Next()
	})

	// Register routes
	routes.RegisterRoutes(app)

	// Start server
	log.Fatal(app.Listen(":4000"))
}
