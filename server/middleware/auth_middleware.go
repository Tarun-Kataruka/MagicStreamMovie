package middleware

import (
	"net/http"

	"github.com/Tarun-Kataruka/MagicStreamMovies/server/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc{
	return func(c *gin.Context){
		token, err := utils.GetAccessToken(c)
		if err != nil{
			c.JSON(401, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		if token == ""{
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no token provided"})
			c.Abort()
			return
		}
		claims, err := utils.ValidateToken(token)
		if err != nil{
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		c.Set("userId", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}