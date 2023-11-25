package auth

import (
	"context"
	"net/http"

	"github.com/viddrobnic/sparovec/models"
)

type contextKey string

const contextKeyUser = contextKey("user")

type Service interface {
	ValidateSession(session *models.Session) error
}

func CreateMiddleware(service Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(models.SessionCookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			session, err := models.SessionFromCookie(cookie.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if err := service.ValidateSession(session); err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyUser, session.User)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequiredMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/auth/sign-in", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func GetUser(r *http.Request) *models.User {
	user := r.Context().Value(contextKeyUser)
	if user == nil {
		return nil
	}

	if user, ok := user.(*models.User); ok {
		return user
	}

	return nil
}
