package middleware

import (
	"art/internal/controllers"
	"art/internal/types"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Middleware func(fn http.HandlerFunc, userControllers *controllers.UserControllers) http.HandlerFunc

func ApplyMiddleware(handler http.HandlerFunc, userControllers *controllers.UserControllers, middlewares ...Middleware) http.HandlerFunc {
	for _, next := range middlewares {
		handler = next(handler, userControllers)
	}
	return handler
}

func Authenticate(next http.HandlerFunc, userControllers *controllers.UserControllers) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		next(w, req)

		isAuthenticated, exists := req.Context().Value(types.CONTEXT_AUTH_KEY).(bool)

		if !exists || !isAuthenticated {
			return
		}

		session := uuid.NewString()

		userControllers.Users.SetAuthenticated(session, true)

		cookie := http.Cookie{
			Name:     types.SESSION_COOKIE,
			Value:    session,
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteNoneMode,
		}

		http.SetCookie(w, &cookie)

		w.Write([]byte("you are logged in"))
	}
}

func Authorize(next http.HandlerFunc, userControllers *controllers.UserControllers) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		cookie, err := req.Cookie(types.SESSION_COOKIE)

		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				http.Error(w, "You are not logged in", http.StatusBadRequest)
			default:
				log.Println(err)
				http.Error(w, "server error", http.StatusInternalServerError)
			}
			return
		}

		if !userControllers.Users.IsAuthenticated(cookie.Value) {
			http.Error(w, "Your session is expired", http.StatusUnauthorized)
			return
		}

		next(w, req)
	}
}
