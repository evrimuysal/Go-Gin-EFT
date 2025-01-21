package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id       primitive.ObjectID `json:"id,omitempty"`
	Name     string             `json:"name,omitempty" validate:"required"`
	Email    string             `json:"email,omitempty" validate:"required,email"`
	Password string             `json:"password,omitempty" validate:"required"`
	Location string             `json:"location,omitempty"`
	Title    string             `json:"title,omitempty"`
}
