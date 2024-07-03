package db

import (
	"art/internal/models"
	"context"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

type MongoGalleryState struct {
	DB *mongo.Database
}

func NewMongoConnection() *mongo.Database {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	mongoURI := os.Getenv("MONGODB")
	clientOptions := options.Client().ApplyURI(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully connected to the database")
	}
	DB := client.Database("gallery")
	return DB
}
func NewMongoGalleryState() *MongoGalleryState {
	return &MongoGalleryState{
		NewMongoConnection(),
	}
}

func (w *MongoGalleryState) List() []models.Painting {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cursor, err := w.DB.Collection("paintings").Find(ctx, bson.M{})
	if err != nil {
		log.Println(err)
		return nil
	}
	defer cursor.Close(ctx)

	var paintings []models.Painting
	if err = cursor.All(ctx, &paintings); err != nil {
		log.Println(err)
		return nil
	}

	return paintings
}

func (w *MongoGalleryState) Save(p models.Painting) primitive.ObjectID {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := w.DB.Collection("paintings").InsertOne(ctx, p)
	if err != nil {
		log.Println(err)
		return primitive.NilObjectID
	}

	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Println("InsertedID is not a valid ObjectID")
		return primitive.NilObjectID
	}

	return id
}

func (w *MongoGalleryState) One(id primitive.ObjectID) models.Painting {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var painting models.Painting
	err := w.DB.Collection("paintings").FindOne(ctx, bson.M{"_id": id}).Decode(&painting)
	if err != nil {
		log.Println(err)
		return models.Painting{}
	}

	return painting
}

func (w *MongoGalleryState) Delete(id primitive.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := w.DB.Collection("paintings").DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (w *MongoGalleryState) Update(id primitive.ObjectID, update bson.M) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := w.DB.Collection("paintings").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
