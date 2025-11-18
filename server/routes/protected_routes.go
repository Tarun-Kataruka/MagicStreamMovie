package routes

import (
	controller "github.com/Tarun-Kataruka/MagicStreamMovies/server/controllers"
	"github.com/gin-gonic/gin"

	verify "github.com/Tarun-Kataruka/MagicStreamMovies/server/middleware"
)

func SetUpProctectedRoutes(router *gin.Engine) {
	router.Use(verify.AuthMiddleware())

	router.GET("/movie/:imdb_id", controller.GetMovie())
	router.POST("/addmovie", controller.AddMovie())
	router.GET("/recommendedmovies", controller.GetRecommendedMovies())
	router.PATCH("/updatemovie/:imdb_id", controller.AdminReviewUpdate())
}
