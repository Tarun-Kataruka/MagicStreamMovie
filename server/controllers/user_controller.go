package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Tarun-Kataruka/MagicStreamMovies/server/database"
	"github.com/Tarun-Kataruka/MagicStreamMovies/server/models"
	"github.com/Tarun-Kataruka/MagicStreamMovies/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection("users")

func HashPassword(password string) (string, error) {
	HashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(HashPassword), err
}

func RegisterUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		err := c.BindJSON(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		validate := validator.New()
		if err = validate.Struct(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", " details": err.Error()})
			return
		}

		hashedPassword, err := HashPassword(user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking for existing user"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
			return
		}

		user.UserID = bson.NewObjectID().Hex()
		user.Password = hashedPassword
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()

		result, err := userCollection.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting user into database"})
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

func LoginUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var userLogin models.UserLogin
		err := c.BindJSON(&userLogin)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var foundUser models.User
		err = userCollection.FindOne(ctx, bson.M{"email": userLogin.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(userLogin.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		token, refreshToken, err := utils.GenerateToken(foundUser.Email, foundUser.FirstName, foundUser.LastName, foundUser.Role, foundUser.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating tokens"})
			return
		}
		if err = utils.UpdateAllTokens(foundUser.UserID, token, refreshToken); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating tokens"})
			return
		}
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "access_token",
			Value:    token,
			Path:   "/",
			MaxAge: 86400,
			Secure: true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "refresh_token",
			Value:    token,
			Path:   "/",
			MaxAge: 604800,
			Secure: true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		c.JSON(http.StatusOK, models.UserResponse{
			UserID:          foundUser.UserID,
			FirstName:       foundUser.FirstName,
			LastName:        foundUser.LastName,
			Email:           foundUser.Email,
			Role:            foundUser.Role,
			//Token:           token,
			//RefreshToken:    refreshToken,
			FavouriteGenres: foundUser.FavouriteGenres,
		})
	}
}

func LogoutHandler() gin.HandlerFunc{
	return func(c *gin.Context){
		var UserLogout struct{
			UserID string `json:"user_id"`
		}
		err := c.ShouldBindJSON(&UserLogout)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}
		fmt.Println("user id from logout handler:", UserLogout.UserID)
		err = utils.UpdateAllTokens(UserLogout.UserID, "", "")
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error logging out user"})
			return
		}
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "access_token",
			Value:    "",
			Path:  "/",
			MaxAge: -1,
			Secure: true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:  "/",
			MaxAge: -1,
			Secure: true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		c.JSON(http.StatusOK, gin.H{"message": "User logged out successfully"})
	}
}

func RefreshTokenHandler() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil{
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not found"})
			return
		}
		claim, err := utils.ValidateRefreshToken(refreshToken)
		if err != nil || claim == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		var user models.User
		err = userCollection.FindOne(ctx, bson.D{{Key:"user_id", Value: claim.UserID}}).Decode(&user)
		if err != nil{
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		newToken, newRefreshToken, _ := utils.GenerateToken(user.Email, user.FirstName, user.LastName, user.Role, user.UserID)
		err = utils.UpdateAllTokens(user.UserID, newToken, newRefreshToken)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating tokens"})
			return
		}
		c.SetCookie("access_token", newToken, 86400, "/", "", true, true)
		c.SetCookie("refresh_token", newRefreshToken, 604800, "/", "", true, true)
		c.JSON(http.StatusOK, gin.H{"message": "Tokens refreshed successfully"})
	}	
}