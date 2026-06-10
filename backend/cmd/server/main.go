package main

import (
	"log"
	"os"

	"script-kill-backend/internal/database"
	"script-kill-backend/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	database.Connect()

	handlers.AutoMigrate()
	handlers.SeedData()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		api.GET("/scripts", handlers.GetScripts)
		api.GET("/carpools", handlers.GetCarpools)
		api.POST("/carpools", handlers.CreateCarpool)
		api.POST("/carpools/join", handlers.JoinCarpool)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	log.Fatal(r.Run(":" + port))
}
