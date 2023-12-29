package auth

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sagikazarmark/slog-shim"
	"github.com/viddrobnic/sparovec/config"
	"github.com/viddrobnic/sparovec/models"
)

type Repository interface {
	GetByUsername(ctx context.Context, username string) (*models.UserCredentials, error)
	Insert(ctx context.Context, username, password, salt string) (*models.UserCredentials, error)
}

type Auth struct {
	repository Repository

	conf *config.Config
	log  *slog.Logger
}

func New(repository Repository, conf *config.Config, log *slog.Logger) *Auth {
	return &Auth{
		repository: repository,
		conf:       conf,
		log:        log,
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
	ctx := r.Context()
	user := GetUser(r)
	if user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	view := signInView(signInViewData{})
	err := view.Render(ctx, w)
	if err != nil {
		a.log.ErrorContext(ctx, "Failed to render sign in template", "error", err)
	}
}

func (a *Auth) submitSignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := a.Authenticate(ctx, username, password)
	if err == models.ErrInvalidCredentials {
		data := signInViewData{
			Username: username,
			Password: password,
			Error:    "Invalid credentials",
		}

		view := signInView(data)
		err := view.Render(ctx, w)
		if err != nil {
			a.log.ErrorContext(ctx, "Failed to render template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	} else if err != nil {
		a.log.ErrorContext(ctx, "Failed to authenticate user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	session, err := a.CreateSession(user)
	if err != nil {
		a.log.ErrorContext(ctx, "Failed to create session", "error", err)
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
