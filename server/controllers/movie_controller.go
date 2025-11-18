package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Tarun-Kataruka/MagicStreamMovies/server/database"
	"github.com/Tarun-Kataruka/MagicStreamMovies/server/models"
	"github.com/Tarun-Kataruka/MagicStreamMovies/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/openai"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var movieCollection *mongo.Collection = database.OpenCollection("movies")
var rankingCollection *mongo.Collection = database.OpenCollection("rankings")
var genreCollection *mongo.Collection = database.OpenCollection("genres")
var validate = validator.New()

func GetMovies() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var movies []models.Movie
		cursor, err := movieCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching movies from database"})
			return
		}
		defer cursor.Close(ctx)
		if err = cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding movies from database"})
			return
		}
		c.JSON(http.StatusOK, movies)
	}
}

func GetMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		movieID := c.Param("imdb_id")
		if movieID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID is required"})
			return
		}
		var movie models.Movie
		err := movieCollection.FindOne(ctx, bson.M{"imdb_id": movieID}).Decode(&movie)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found in database"})
			return
		}
		c.JSON(http.StatusOK, movie)
	}
}

func AddMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var movie models.Movie
		if err := c.BindJSON(&movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie data"})
			return
		}
		if err := validate.Struct(movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := movieCollection.InsertOne(ctx, movie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting movie into database"})
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

func AdminReviewUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := utils.GetRoleFromContext(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role not found in context"})
			return
		}
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update movie reviews"})
			return
		}
		movieId := c.Param("imdb_id")
		if movieId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID is required"})
			return
		}
		var req struct {
			AdminReview string `json:"admin_review"`
		}
		var resp struct {
			RankingName string `json:"ranking_name"`
			AdminReview string `json:"admin_review"`
		}
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}
		sentiment, rankVal, err := GetReviewRanking(req.AdminReview)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting review ranking"})
			return
		}
		filter := bson.M{"imdb_id": movieId}
		update := bson.M{"$set": bson.M{
			"admin_review": req.AdminReview,
			"ranking": bson.M{
				"ranking_value": rankVal,
				"ranking_name":  sentiment,
			},
		}}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		result, err := movieCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating movie review"})
			return
		}
		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}
		resp.RankingName = sentiment
		resp.AdminReview = req.AdminReview
		c.JSON(http.StatusOK, resp)
	}
}

func GetReviewRanking(admin_review string) (string, int, error) {
	rankings, err := GetRanking()
	if err != nil {
		return "", 0, err
	}
	sentimentDelimited := ""
	for _, ranking := range rankings {
		if ranking.RankingValue != 999 {
			sentimentDelimited = sentimentDelimited + ranking.RankingName + ","
		}
	}
	sentimentDelimited = strings.Trim(sentimentDelimited, ",")
	err = godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: No .env file found")
	}
	OpenaiApiKey := os.Getenv("OPENAI_API_KEY")
	if OpenaiApiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}
	llm, err := openai.New()
	if err != nil {
		return "", 0, err
	}
	base_prompt := os.Getenv("BASE_PROMPT_TEMPLATE")
	if base_prompt == "" {
		log.Fatal("BASE_PROMPT_TEMPLATE environment variable not set")
	}
	prompt := strings.Replace(base_prompt, "{rankings}", sentimentDelimited, 1)
	response, err := llm.Call(context.Background(), prompt+admin_review)
	if err != nil {
		return "", 0, err
	}
	rankVal := 0
	for _, ranking := range rankings {
		if ranking.RankingName == response {
			rankVal = ranking.RankingValue
			break
		}
	}
	return response, rankVal, nil
}

func GetRanking() ([]models.Ranking, error) {
	var rankings []models.Ranking
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	cursor, err := rankingCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &rankings); err != nil {
		return nil, err
	}
	return rankings, nil
}

func GetRecommendedMovies() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		favGenres, err := GetUserFavouriteGenres(userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = godotenv.Load(".env")
		if err != nil {
			log.Println("Warning: No .env file found")
		}
		var recommendedMovieLimitVal int64
		recommendedMovieLimit := os.Getenv("RECOMMENDED_MOVIE_LIMIT")
		if recommendedMovieLimit != "" {
			recommendedMovieLimitVal, _ = strconv.ParseInt(recommendedMovieLimit, 10, 64)
		}
		findOptions := options.Find()
		findOptions.SetSort(bson.D{{Key: "ranking.ranking_value", Value: 1}})
		filter := bson.M{"genre.genre_name": bson.M{"$in": favGenres}}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		cursor, err := movieCollection.Find(ctx, filter, findOptions.SetLimit(recommendedMovieLimitVal))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching recommended movies"})
			return
		}
		defer cursor.Close(ctx)
		var recommendedMovies []models.Movie
		if err = cursor.All(ctx, &recommendedMovies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding recommended movies"})
			return
		}
		c.JSON(http.StatusOK, recommendedMovies)
	}
}

func GetUserFavouriteGenres(userId string) ([]string, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	filter := bson.M{"user_id": userId}
	projection := bson.M{
		"favourite_genres.genre_name": 1,
		"_id":                         0,
	}
	opts := options.FindOne().SetProjection(projection)
	var results bson.M
	err := userCollection.FindOne(ctx, filter, opts).Decode(&results)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []string{}, nil
		}
		return nil, err
	}
	favGenresArray, ok := results["favourite_genres"].(bson.A)
	if !ok {
		return []string{}, errors.New("unable to retrieve favourite genres")
	}
	var genreNames []string
	for _, item := range favGenresArray {
		if genreMap, ok := item.(bson.D); ok {
			for _, elem := range genreMap {
				if elem.Key == "genre_name" {
					if name, ok := elem.Value.(string); ok {
						genreNames = append(genreNames, name)
					}
				}
			}
		}
	}
	return genreNames, nil
}

func GetGenres() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var genres []models.Genre
		cursor, err := genreCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching genres from database"})
			return
		}
		defer cursor.Close(ctx)
		if err := cursor.All(ctx, &genres); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding genres from database"})
			return
		}
		c.JSON(http.StatusOK, genres)
	}
}
