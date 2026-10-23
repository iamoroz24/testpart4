package main

import (
	"log"
	"project03/internal/di"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	app := di.BuildContainer()
	app.Run()
}
