package utils

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/Tarun-Kataruka/MagicStreamMovies/server/database"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	Role      string
	UserID    string
	jwt.RegisteredClaims
}

var SECRET_KEY string = os.Getenv("SECRET_KEY")
var REFRESH_SECRET_KEY string = os.Getenv("SECRET_REFRESH_KEY")
var userCollection *mongo.Collection = database.OpenCollection("users")

func GenerateToken(email, firstName, lastName, role, userId string) (string, string, error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		UserID:    userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "MagicStreamMovies",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}
	refreshClaims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		UserID:    userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "MagicStreamMovies",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(168 * time.Hour)), // 7 days
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(REFRESH_SECRET_KEY))
	if err != nil {
		return "", "", err
	}
	return signedToken, signedRefreshToken, nil
}

func UpdateAllTokens(userId, token, refreshToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateData := bson.M{
		"$set": bson.M{
			"token":         token,
			"refresh_token": refreshToken,
			"updated_at":    updatedAt,
		},
	}
	err := userCollection.FindOneAndUpdate(ctx, bson.M{"user_id": userId}, updateData).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetAccessToken(c *gin.Context) (string, error) {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}
	tokenString := authHeader[len("Bearer "):]
	if tokenString == "" {
		return "", errors.New("token missing in authorization header")
	}
	return tokenString, nil
}

func ValidateToken(signedToken string) (*SignedDetails, error) {
	claims := &SignedDetails{}
	token, err := jwt.ParseWithClaims(signedToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return nil, err
	}
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, err
	}
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	return claims, nil
}

func GetUserIdFromContext(c *gin.Context) (string, error) {
	userId, exists := c.Get("userId")
	if !exists {
		return "", errors.New("userId not found in context")
	}
	id, ok := userId.(string)
	if !ok {
		return "", errors.New("Unable to retrieve userId")
	}
	return id, nil
}


func GetRoleFromContext(c *gin.Context) (string, error) {
	role, exists := c.Get("role")
	if !exists {
		return "", errors.New("userId not found in context")
	}
	memberRole, ok := role.(string)
	if !ok {
		return "", errors.New("Unable to retrieve userId")
	}
	return memberRole, nil
}
