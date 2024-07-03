package controllers

import (
	"art/internal/drive"
	models "art/internal/models"
	"art/internal/photoprocessor"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type GalleryController struct {
	Gallery *models.Gallery
}

type Response struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (w *GalleryController) ListProducts(res http.ResponseWriter, req *http.Request) {
	products := w.Gallery.ListProducts()

	res.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(res).Encode(products)
	if err != nil {
		return
	}
}

func (w *GalleryController) GetOnePainting(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	painting := w.Gallery.GetOnePainting(id)
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(painting)
}

func (w *GalleryController) AddPainting(res http.ResponseWriter, req *http.Request) {
	err := req.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	files := req.MultipartForm.File["images"]
	processor := &photoprocessor.LocalPhotoProcessor{}
	err = photoprocessor.SaveAndResizeFiles(processor, files, "tmp/tmpFiles", "tmp/resizedFiles")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	validation := NewValidation(req)
	painting, err := validation.Validate("price", "date", "materials", "size", "title", "titleUkr", "description", "descriptionUkr", "availability")

	painting.Photos, err = drive.UploadImages(req.FormValue("title"), "tmp/resizedFiles")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	id := w.Gallery.AddProduct(painting)

	fmt.Printf("Inserted painting with ID: %s\n", id.Hex())

	defer func() {
		err = processor.RemoveFolder("tmp/tmpFiles")
		if err != nil {
			log.Println(err)
		}
		err = processor.RemoveFolder("tmp/resizedFiles")
		if err != nil {
			log.Println(err)
		}
	}()

	response := Response{
		Data:    id.Hex(),
		Message: "Painting created successfully",
	}

	res.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(res).Encode(response)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (w *GalleryController) DeletePainting(res http.ResponseWriter, req *http.Request) {
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := mux.Vars(req)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	painting := w.Gallery.GetOnePainting(id)
	err = drive.DeleteFolder(painting.Photos.FolderId)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if !w.Gallery.DeletePainting(id) {
		http.Error(res, "Painting not found", http.StatusNotFound)
		return
	}

	fmt.Printf("Deleted painting with ID: %s\n", id.Hex())
	res.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(res, "Painting deleted successfully")
}

func (w *GalleryController) UpdatePainting(res http.ResponseWriter, req *http.Request) {
	err := req.ParseMultipartForm(10 << 20) //
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	files := req.MultipartForm.File["images"]

	params := mux.Vars(req)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	update := bson.M{}
	// If there are new images, delete the old folder from Google Drive

	if len(files) > 0 {
		files := req.MultipartForm.File["images"]
		processor := &photoprocessor.LocalPhotoProcessor{}
		err = photoprocessor.SaveAndResizeFiles(processor, files, "tmp/tmpFiles", "tmp/resizedFiles")
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		err = drive.DeleteFolder(id.Hex())
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		imageLinks, err := drive.UploadImages(id.Hex(), "tmp/tmpFiles")
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		update["photos"] = imageLinks
		defer func() {
			err = processor.RemoveFolder("tmp/tmpFiles")
			if err != nil {
				log.Println(err)
			}
			err = processor.RemoveFolder("tmp/resizedFiles")
			if err != nil {
				log.Println(err)
			}
		}()

	}

	if req.FormValue("title") != "" {
		update["title"] = req.FormValue("title")
	}
	if req.FormValue("titleUkr") != "" {
		update["titleUkr"] = req.FormValue("titleUkr")
	}
	if req.FormValue("description") != "" {
		update["description"] = req.FormValue("description")
	}
	if req.FormValue("descriptionUkr") != "" {
		update["descriptionUkr"] = req.FormValue("descriptionUkr")
	}
	if priceStr := req.FormValue("price"); priceStr != "" {
		price, err := strconv.ParseFloat(priceStr, 64)
		if err == nil {
			update["price"] = price
		}
	}
	if dateStr := req.FormValue("date"); dateStr != "" {
		date, err := time.Parse(time.RFC3339, dateStr)
		if err == nil {
			update["date"] = primitive.NewDateTimeFromTime(date)
		}
	}
	if materialsStr := req.FormValue("materials"); materialsStr != "" {
		var materials []models.Material
		err := json.Unmarshal([]byte(materialsStr), &materials)
		if err == nil {
			update["materials"] = materials
		}
	}
	if sizeStr := req.FormValue("size"); sizeStr != "" {
		var size []interface{}
		err := json.Unmarshal([]byte(sizeStr), &size)
		if err == nil {
			update["size"] = size
		}
	}
	if req.FormValue("availability") != "" {
		update["availability"] = req.FormValue("availability")
	}

	if len(update) > 0 {
		if !w.Gallery.UpdatePainting(id, bson.M{"$set": update}) {
			http.Error(res, "Failed to update painting", http.StatusInternalServerError)
			return
		}
	}
	_, _ = io.WriteString(res, "Painting updated successfully")
}
