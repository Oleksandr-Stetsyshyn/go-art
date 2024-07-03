package controllers

import (
	"art/internal/models"
	"art/internal/types"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type attempt struct {
	Login    types.Login    `json:"Login"`
	Password types.Password `json:"Password"`
}

type UserControllers struct {
	Users *models.Users
}

func (services *UserControllers) Register(w http.ResponseWriter, r *http.Request) {
	log.Printf("Register handler")
	body, err := io.ReadAll(r.Body)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("internal error"))
		return
	}

	var newAttempt attempt

	err = json.Unmarshal(body, &newAttempt)

	if err != nil {
		log.Println(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if newAttempt.Login == "" || newAttempt.Password == "" {
		http.Error(w, "Login and Password must be present", http.StatusBadRequest)
		return
	}

	h := sha1.New()
	h.Write([]byte(newAttempt.Password))
	hashedPassword := types.Password(hex.EncodeToString(h.Sum(nil)))

	user := models.User{
		Login:    newAttempt.Login,
		Password: hashedPassword,
	}

	err = services.Users.Register(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Registered a new user: %s\n", newAttempt.Login)

	w.Write([]byte("you are registered"))
}

func (services *UserControllers) SignIn(w http.ResponseWriter, r *http.Request) {
	log.Printf("SignIn handler")
	body, err := io.ReadAll(r.Body)

	if err != nil {
		log.Println(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var newAttempt attempt

	err = json.Unmarshal(body, &newAttempt)

	if err != nil {
		log.Println(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h := sha1.New()
	h.Write([]byte(newAttempt.Password))
	hashedPassword := types.Password(hex.EncodeToString(h.Sum(nil)))

	user := models.User{
		Login:    newAttempt.Login,
		Password: hashedPassword,
	}

	user, err = services.Users.Login(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	services.Users.SetAuthenticated(string(newAttempt.Login), true)
	*r = *r.Clone(context.WithValue(r.Context(), types.CONTEXT_AUTH_KEY, true))
}
