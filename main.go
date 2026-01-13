package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/lanxizhu/go-playground/utils"

	// _ "github.com/joho/godotenv/autoload"
	"github.com/lanxizhu/go-playground/router"
)

var ctx = context.Background()

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using default environment variables")
	} else {
		log.Println(".env file loaded successfully")
	}

	port := os.Getenv("PORT")
	log.Println("PORT: " + port)

	r := router.SetupRouter()

	utils.SetupRedis(ctx)

	err := r.Run()
	if err != nil {
		return
	} // listens on 0.0.0.0:8080 by default
}
