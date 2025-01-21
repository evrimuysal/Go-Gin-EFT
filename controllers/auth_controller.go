package controllers

import (
	"context"
	"net/http"
	"time"

	"gin-mongo-api/configs"
	"gin-mongo-api/models"
	"gin-mongo-api/responses"
	"gin-mongo-api/validators"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

const (
	passwordHashCost = bcrypt.DefaultCost
	requestTimeout   = 10 * time.Second
	successMessage   = "success"
	errorMessage     = "error"
	jwtSecret        = "your_secret_key"
	jwtExpiration    = 24 * time.Hour
)

func respondWithJSON(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, responses.UserResponse{
		Status:  status,
		Message: message,
		Data:    map[string]interface{}{"data": data},
	})
}

func generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(jwtExpiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtSecret))
}

func Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			respondWithJSON(c, http.StatusBadRequest, errorMessage, err.Error())
			return
		}
		if validationErr := validators.Validate.Struct(&user); validationErr != nil {
			respondWithJSON(c, http.StatusBadRequest, errorMessage, validationErr.Error())
			return
		}

		hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(user.Password), passwordHashCost)
		if hashErr != nil {
			respondWithJSON(c, http.StatusInternalServerError, errorMessage, hashErr.Error())
			return
		}
		user.Password = string(hashedPassword)

		user.Id = primitive.NewObjectID()

		if _, dbErr := userCollection.InsertOne(ctx, user); dbErr != nil {
			respondWithJSON(c, http.StatusInternalServerError, errorMessage, dbErr.Error())
			return
		}

		respondWithJSON(c, http.StatusCreated, successMessage, "User registered successfully")
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			respondWithJSON(c, http.StatusBadRequest, errorMessage, err.Error())
			return
		}
		configs.UserCollection = configs.GetCollection(configs.DB, "users")
		dbErr := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if dbErr != nil {
			respondWithJSON(c, http.StatusUnauthorized, errorMessage, "Invalid email or password")
			return
		}

		if compareErr := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(user.Password)); compareErr != nil {
			respondWithJSON(c, http.StatusUnauthorized, errorMessage, "Invalid email or password")
			return
		}

		token, tokenErr := generateToken(foundUser.Id.Hex())
		if tokenErr != nil {
			respondWithJSON(c, http.StatusInternalServerError, errorMessage, "Failed to generate token")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  successMessage,
			"message": "Login successful",
			"token":   token,
		})
	}
}
