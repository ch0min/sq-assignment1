package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"github.com/oTuff/sq-ola1/db"
	_ "github.com/oTuff/sq-ola1/docs"
	"github.com/oTuff/sq-ola1/routes"
)

func main() {
	database, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Initialize Fiber app
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Get("/swagger/*", swagger.HandlerDefault) // default

	// Setup routes
	routes.SetupTodoRoutes(app, database)

	// Start the Fiber app on port 4000
	log.Fatal(app.Listen(":4000"))
}
