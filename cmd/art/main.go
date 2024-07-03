package main

import (
	"art/internal/api"
	"art/internal/controllers"
	"art/internal/db"
	users "art/internal/models"
	"net/http"
)

func main() {
	mongo := db.NewMongoGalleryState()
	mongoUsers := db.NewMongoUserState(mongo.DB)

	gl := users.NewGallery(mongo)
	glc := &controllers.GalleryController{Gallery: gl}

	us := users.NewUsers(mongoUsers)
	usc := &controllers.UserControllers{Users: us}

	r := api.NewRouter(glc, usc)
	err := http.ListenAndServe("localhost:8080", r)
	if err != nil {
		return
	}
}
