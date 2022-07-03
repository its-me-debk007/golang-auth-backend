package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/its-me-debk007/golang-auth-backend/database"
	"github.com/its-me-debk007/golang-auth-backend/routes"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	database.ConnectDB()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowCredentials: true,
	}))

	routes.SetupRoutes(app)

	if err := godotenv.Load(); err != nil {
		log.Fatal(err.Error())
	}

	port := os.Getenv("PORT")

	log.Println("Port is " + port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatal("App listen error:-\n" + err.Error())
	}
}
