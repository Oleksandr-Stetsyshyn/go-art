package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Material struct {
	ID  string `bson:"id,omitempty"`
	EN  string `bson:"en,omitempty"`
	UKR string `bson:"ukr,omitempty"`
}

type Photos struct {
	Urls     []string `bson:"images,omitempty"`
	FolderId string   `bson:"folderId,omitempty"`
}

type Painting struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Images         []string           `bson:"images,omitempty" json:"images"`
	Photos         Photos             `bson:"photos,omitempty" json:"photos"`
	Title          string             `bson:"title,omitempty" json:"title"`
	TitleUkr       string             `bson:"titleUkr,omitempty" json:"titleUkr"`
	Description    string             `bson:"description,omitempty" json:"description"`
	DescriptionUkr string             `bson:"descriptionUkr,omitempty" json:"descriptionUkr"`
	Price          float64            `bson:"price,omitempty" json:"price"`
	Size           []interface{}      `bson:"size,omitempty" json:"size"`
	Date           primitive.DateTime `bson:"date,omitempty" json:"date"`
	Availability   string             `bson:"availability,omitempty" json:"availability"`
	Materials      []Material         `bson:"materials,omitempty" json:"materials"`
}
