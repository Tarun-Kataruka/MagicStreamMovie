package main

import (
	"fmt"
	"time"

	"github.com/Tarun-Kataruka/MagicStreamMovies/server/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	config := cors.Config{}
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.ExposeHeaders = []string{"COntent-Length"}
	config.MaxAge = 12 * time.Hour
	router.Use(cors.New(config))

	routes.SetUpUnProctectedRoutes(router)
	routes.SetUpProctectedRoutes(router)

	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
