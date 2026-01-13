package router

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lanxizhu/go-playground/upload"
)

func SetupRouter() *gin.Engine {
	gin.ForceConsoleColor()
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	r.Use(cors.New(config))

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	dir, err := os.Getwd()

	if err != nil {
		log.Fatal("Failed to get current working directory:", err)
	}
	fmt.Println("Current working directory:", dir)

	r.POST("upload", upload.Upload)
	r.GET("upload/status", upload.Status)
	r.POST("upload/check", upload.Chunk)
	r.POST("upload/complete", upload.Complete)

	return r
}
