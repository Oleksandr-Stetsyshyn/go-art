package api

import (
	"art/internal/controllers"
	"art/internal/middleware"
	"github.com/gorilla/mux"
)

func NewRouter(glc *controllers.GalleryController, usc *controllers.UserControllers) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/register", usc.Register).Methods("POST")
	r.HandleFunc("/login", middleware.ApplyMiddleware(usc.SignIn, usc, middleware.Authenticate)).Methods("POST")

	r.HandleFunc("/paintings", glc.ListProducts).Methods("GET")
	//r.HandleFunc("/paintings", middleware.Authorize(glc.ListProducts, usc)).Methods("GET")

	r.HandleFunc("/paintings/add", glc.AddPainting).Methods("POST")

	r.HandleFunc("/paintings/{id}", glc.GetOnePainting).Methods("GET")
	r.HandleFunc("/paintings/{id}", glc.DeletePainting).Methods("DELETE")
	r.HandleFunc("/paintings/{id}", glc.UpdatePainting).Methods("PUT")

	return r
}
