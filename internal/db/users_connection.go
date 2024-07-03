package db

import (
	"art/internal/models"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type MongoUserState struct {
	DB            *mongo.Database
	Authenticated map[string]bool
}

func NewMongoUserState(db *mongo.Database) *MongoUserState {
	return &MongoUserState{
		DB:            db,
		Authenticated: make(map[string]bool),
	}
}

func (u *MongoUserState) Register(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var result models.User
	err := u.DB.Collection("users").FindOne(ctx, bson.M{"login": user.Login}).Decode(&result)
	if err != mongo.ErrNoDocuments {
		return errors.New("user already exists")
	}

	insertResult, err := u.DB.Collection("users").InsertOne(ctx, user)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("Inserted a new user with ID: %v", insertResult.InsertedID)

	return nil
}

func (u *MongoUserState) Login(user models.User) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var result models.User
	err := u.DB.Collection("users").FindOne(ctx, bson.M{"login": user.Login}).Decode(&result)
	if err != nil {
		log.Println(err)
		return models.User{}, err
	}

	if result.Password != user.Password {
		return models.User{}, errors.New("invalid password")
	}

	return result, nil
}

func (u *MongoUserState) SetAuthenticated(session string, authenticated bool) {
	u.Authenticated[session] = authenticated
}

func (u *MongoUserState) IsAuthenticated(session string) bool {
	authenticated, exists := u.Authenticated[session]
	return exists && authenticated
}
