package models

import (
	"art/internal/types"
	"crypto/sha1"
	"encoding/hex"
)

type User struct {
	Login    types.Login
	Password types.Password
}

type UserState interface {
	Register(user User) error
	Login(user User) (User, error)
	SetAuthenticated(session string, authenticated bool)
	IsAuthenticated(session string) bool
}

type Users struct {
	state         UserState
	Registered    map[types.Login]types.Password
	Authenticated map[string]bool
}

func NewUsers(state UserState) *Users {
	return &Users{
		state:         state,
		Registered:    make(map[types.Login]types.Password),
		Authenticated: make(map[string]bool),
	}
}

func (w *Users) Register(user User) error {
	h := sha1.New()
	h.Write([]byte(user.Password))
	user.Password = types.Password(hex.EncodeToString(h.Sum(nil)))

	return w.state.Register(user)
}

func (w *Users) Login(user User) (User, error) {
	h := sha1.New()
	h.Write([]byte(user.Password))
	user.Password = types.Password(hex.EncodeToString(h.Sum(nil)))

	return w.state.Login(user)
}

func (w *Users) SetAuthenticated(session string, authenticated bool) {
	w.Authenticated[session] = authenticated
}

func (w *Users) IsAuthenticated(session string) bool {
	authenticated, exists := w.Authenticated[session]
	return exists && authenticated
}
