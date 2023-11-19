package routes

import (
	"context"
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/viddrobnic/sparovec/models"
)

type AuthService interface {
	Authenticate(ctx context.Context, username, password string) (*models.User, error)
	CreateSession(user *models.User) (*models.Session, error)
}

type Auth struct {
	service AuthService

	// Templates
	signInTemplate *template.Template
}

func NewAuth(service AuthService) *Auth {
	signInTemplate := template.Must(template.ParseFiles(
		"templates/index.html",
		"templates/auth/sign-in.html",
	))

	return &Auth{
		service: service,

		signInTemplate: signInTemplate,
	}
}

func (a *Auth) Mount(group *echo.Group) {
	group.GET("/sign-in", a.signIn)
	group.POST("/sign-in", a.submitSignIn)

	group.GET("/sign-out", a.signOut)
}

func (a *Auth) signIn(c echo.Context) error {
	user, _ := c.Get(models.UserContextKey).(*models.User)
	if user != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	c.Response().WriteHeader(http.StatusOK)
	return a.signInTemplate.Execute(c.Response().Writer, nil)
}

type submitRequest struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

func (a *Auth) submitSignIn(c echo.Context) error {
	var req submitRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Bad request")
	}

	user, err := a.service.Authenticate(c.Request().Context(), req.Username, req.Password)
	if err == models.ErrInvalidCredentials {
		data := &models.SignInContext{
			Username: req.Username,
			Password: req.Password,
			Error:    "Invalid credentials",
		}
		return a.signInTemplate.Execute(c.Response().Writer, data)
	} else if err != nil {
		return err
	}

	session, err := a.service.CreateSession(user)
	if err != nil {
		return err
	}

	sessionCookie, err := session.ToCookie()
	if err != nil {
		return err
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
	c.SetCookie(cookie)

	return c.Redirect(http.StatusSeeOther, "/")
}

func (a *Auth) signOut(c echo.Context) error {
	cookie := &http.Cookie{
		Name:     models.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	return c.Redirect(http.StatusSeeOther, "/auth/sign-in")
}
