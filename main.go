package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/lanxizhu/go-playground/utils"
	"go.uber.org/zap"

	// _ "github.com/joho/godotenv/autoload"
	"github.com/lanxizhu/go-playground/router"
)

var logger *zap.Logger

var ctx = context.Background()

func init() {
	// You can perform more advanced configuration here,
	// such as outputting to files, using JSON format, etc.
	var err error
	if gin.Mode() == gin.ReleaseMode {
		// Recommended configuration for production environment — outputs JSON format
		logger, err = zap.NewProduction()
	} else {
		// Recommended configuration for development environment; outputs human readable format
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
}

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

	logger.Info("Application starting up", zap.String("port", os.Getenv("PORT")))

	err := r.Run()
	if err != nil {
		return
	} // listens on 0.0.0.0:8080 by default

	// Delaying `Sync()` ensures all buffered logs have been written.
	defer func(logger *zap.Logger) {
		err = logger.Sync()
		if err != nil {
			log.Fatalf("logger shutdown: %v", err)
		}
	}(logger) // Note: Calling `defer logger.Sync()` in the `main` function is more appropriate to ensure it executes before the program exits.
}
