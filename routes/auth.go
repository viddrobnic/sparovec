package routes

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/viddrobnic/sparovec/middleware/auth"
	"github.com/viddrobnic/sparovec/models"
)

type AuthService interface {
	Authenticate(ctx context.Context, username, password string) (*models.User, error)
	CreateSession(user *models.User) (*models.Session, error)
}

type Auth struct {
	service AuthService
	log     *slog.Logger

	// Templates
	signInTemplate *template.Template
}

func NewAuth(service AuthService, log *slog.Logger) *Auth {
	signInTemplate := template.Must(template.ParseFiles(
		"templates/index.html",
		"templates/auth/sign-in.html",
	))

	return &Auth{
		service: service,
		log:     log,

		signInTemplate: signInTemplate,
	}
}

func (a *Auth) Mount(router chi.Router) {
	group := chi.NewRouter()

	group.Get("/sign-in", a.signIn)
	group.Post("/sign-in", a.submitSignIn)
	group.Get("/sign-out", a.signOut)

	router.Mount("/auth", group)
}

func (a *Auth) signIn(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err := renderTemplate(w, a.signInTemplate, nil)
	if err != nil {
		a.log.Error("Failed to render template", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (a *Auth) submitSignIn(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := a.service.Authenticate(r.Context(), username, password)
	if err == models.ErrInvalidCredentials {
		data := &models.SignInContext{
			Username: username,
			Password: password,
			Error:    "Invalid credentials",
		}
		err := a.signInTemplate.Execute(w, data)
		if err != nil {
			a.log.Error("Failed to render template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	} else if err != nil {
		a.log.Error("Failed to authenticate user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	session, err := a.service.CreateSession(user)
	if err != nil {
		a.log.Error("Failed to create session", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	sessionCookie, err := session.ToCookie()
	if err != nil {
		a.log.Error("Failed to serialize session", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	cookie := &http.Cookie{
		Name:     models.SessionCookieName,
		Value:    sessionCookie,
		Path:     "/",
		Expires:  session.ExpiresAt,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *Auth) signOut(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     models.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/auth/sign-in", http.StatusSeeOther)
}
