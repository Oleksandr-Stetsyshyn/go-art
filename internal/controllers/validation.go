package controllers

import (
	"art/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Validation struct {
	req *http.Request
}

func NewValidation(req *http.Request) *Validation {
	return &Validation{req: req}
}

func (v *Validation) Validate(fields ...string) (models.Painting, error) {
	var painting models.Painting

	for _, field := range fields {
		switch field {
		case "price":
			err := v.validatePrice(&painting)
			if err != nil {
				return models.Painting{}, err
			}
		case "date":
			err := v.validateDate(&painting)
			if err != nil {
				return models.Painting{}, err
			}
		case "materials":
			err := v.validateMaterials(&painting)
			if err != nil {
				return models.Painting{}, err
			}
		case "size":
			err := v.validateSize(&painting)
			if err != nil {
				return models.Painting{}, err
			}
		case "title":
			painting.Title = v.req.FormValue("title")
		case "titleUkr":
			painting.TitleUkr = v.req.FormValue("titleUkr")
		case "description":
			painting.Description = v.req.FormValue("description")
		case "descriptionUkr":
			painting.DescriptionUkr = v.req.FormValue("descriptionUkr")
		case "availability":
			painting.Availability = v.req.FormValue("availability")
		}
	}

	return painting, nil
}

func (v *Validation) validatePrice(painting *models.Painting) error {
	priceStr := v.req.FormValue("price")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return fmt.Errorf("Invalid price value: %v", err)
	}
	painting.Price = price
	return nil
}

func (v *Validation) validateDate(painting *models.Painting) error {
	dateStr := v.req.FormValue("date")
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return fmt.Errorf("Invalid date value: %v", err)
	}
	painting.Date = primitive.NewDateTimeFromTime(date)
	return nil
}

func (v *Validation) validateMaterials(painting *models.Painting) error {
	materialsStr := v.req.FormValue("materials")
	err := json.Unmarshal([]byte(materialsStr), &painting.Materials)
	if err != nil {
		return fmt.Errorf("Invalid materials value: %v", err)
	}
	return nil
}

func (v *Validation) validateSize(painting *models.Painting) error {
	sizeStr := v.req.FormValue("size")
	err := json.Unmarshal([]byte(sizeStr), &painting.Size)
	if err != nil {
		return fmt.Errorf("Invalid size value: %v", err)
	}
	return nil
}
