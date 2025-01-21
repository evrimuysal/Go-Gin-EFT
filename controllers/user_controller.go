package controllers

import (
	"context"
	"gin-mongo-api/configs"
	"gin-mongo-api/models"
	"gin-mongo-api/responses"
	"gin-mongo-api/validators"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	userCollection  *mongo.Collection = configs.GetCollection(configs.DB, "users")
	timeoutDuration                   = 10 * time.Second
)

func sendResponse(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, responses.UserResponse{
		Status:  status,
		Message: message,
		Data:    map[string]interface{}{"data": data},
	})
}

func getObjectIDFromParam(c *gin.Context, paramName string) (primitive.ObjectID, error) {
	id := c.Param(paramName)
	return primitive.ObjectIDFromHex(id)
}

func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			sendResponse(c, http.StatusBadRequest, "error", err.Error())
			return
		}

		if validationErr := validators.Validate.Struct(&user); validationErr != nil {
			sendResponse(c, http.StatusBadRequest, "error", validationErr.Error())
			return
		}

		newUser := models.User{
			Id:       primitive.NewObjectID(),
			Name:     user.Name,
			Location: user.Location,
			Title:    user.Title,
			Email:    user.Email,
			Password: user.Password,
		}

		result, err := userCollection.InsertOne(ctx, newUser)
		if err != nil {
			sendResponse(c, http.StatusInternalServerError, "error", err.Error())
			return
		}

		sendResponse(c, http.StatusCreated, "success", result)
	}
}

func GetAUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()

		objId, err := getObjectIDFromParam(c, "userId")
		if err != nil {
			sendResponse(c, http.StatusBadRequest, "error", "Invalid user ID")
			return
		}

		var user models.User
		if err = userCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&user); err != nil {
			sendResponse(c, http.StatusInternalServerError, "error", err.Error())
			return
		}

		sendResponse(c, http.StatusOK, "success", user)
	}
}

func EditAUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()

		objId, err := getObjectIDFromParam(c, "userId")
		if err != nil {
			sendResponse(c, http.StatusBadRequest, "error", "Invalid user ID")
			return
		}

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			sendResponse(c, http.StatusBadRequest, "error", err.Error())
			return
		}

		if validationErr := validators.Validate.Struct(&user); validationErr != nil {
			sendResponse(c, http.StatusBadRequest, "error", validationErr.Error())
			return
		}

		update := bson.M{"name": user.Name, "location": user.Location, "title": user.Title}
		result, err := userCollection.UpdateOne(ctx, bson.M{"id": objId}, bson.M{"$set": update})
		if err != nil {
			sendResponse(c, http.StatusInternalServerError, "error", err.Error())
			return
		}

		var updatedUser models.User
		if result.MatchedCount == 1 {
			if err := userCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&updatedUser); err != nil {
				sendResponse(c, http.StatusInternalServerError, "error", err.Error())
				return
			}
		}

		sendResponse(c, http.StatusOK, "success", updatedUser)
	}
}

func DeleteAUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()

		objId, err := getObjectIDFromParam(c, "userId")
		if err != nil {
			sendResponse(c, http.StatusBadRequest, "error", "Invalid user ID")
			return
		}

		result, err := userCollection.DeleteOne(ctx, bson.M{"id": objId})
		if err != nil {
			sendResponse(c, http.StatusInternalServerError, "error", err.Error())
			return
		}

		if result.DeletedCount < 1 {
			sendResponse(c, http.StatusNotFound, "error", "User with specified ID not found!")
			return
		}

		sendResponse(c, http.StatusOK, "success", "User successfully deleted!")
	}
}

func GetAllUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()

		var users []models.User
		results, err := userCollection.Find(ctx, bson.M{})
		if err != nil {
			sendResponse(c, http.StatusInternalServerError, "error", err.Error())
			return
		}
		defer results.Close(ctx)

		for results.Next(ctx) {
			var userResult models.User
			if err := results.Decode(&userResult); err != nil {
				sendResponse(c, http.StatusInternalServerError, "error", err.Error())
				return
			}
			users = append(users, userResult)
		}

		sendResponse(c, http.StatusOK, "success", users)
	}
}
