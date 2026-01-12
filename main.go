package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	// _ "github.com/joho/godotenv/autoload"
	"github.com/lanxizhu/go-playground/router"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using default environment variables")
	} else {
		log.Println(".env file loaded successfully")
	}

	port := os.Getenv("PORT")
	log.Println("PORT: " + port)

	r := router.SetupRouter()

	err := r.Run()
	if err != nil {
		return
	} // listens on 0.0.0.0:8080 by default
}
