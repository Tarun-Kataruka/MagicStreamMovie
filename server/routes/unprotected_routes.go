package routes

import (
	controller "github.com/Tarun-Kataruka/MagicStreamMovies/server/controllers"
	"github.com/gin-gonic/gin"
)

func SetUpUnProctectedRoutes(router *gin.Engine) {

	router.POST("/register", controller.RegisterUser())
	router.POST("/login", controller.LoginUser())
	router.POST("/logout", controller. LogoutHandler())
	router.GET("/movies", controller.GetMovies())
	router.GET("/genres", controller.GetGenres())
	router.POST("/refresh", controller.RefreshTokenHandler())
}
