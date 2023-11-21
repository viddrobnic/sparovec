package routes

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/viddrobnic/sparovec/middleware/auth"
	"github.com/viddrobnic/sparovec/models"
)

type TagsService interface {
	List(ctx context.Context, walletId int, user *models.User) ([]*models.Tag, error)
	Create(ctx context.Context, walletId int, name string, user *models.User) (*models.Tag, error)
	Update(ctx context.Context, tagId int, name string, user *models.User) (*models.Tag, error)
	Delete(ctx context.Context, tagId int, user *models.User) error
}

type Tags struct {
	navbarService NavbarWalletsService
	tagsService   TagsService
	log           *slog.Logger

	// Templates
	tagsTemplate *template.Template
}

func NewTags(
	navbarService NavbarWalletsService,
	tagsService TagsService,
	log *slog.Logger,
) *Tags {
	tagsTemplate := template.Must(template.ParseFiles(
		"templates/index.html",
		"templates/layout.html",
		"templates/tags/tags.html",
	))
	template.Must(tagsTemplate.ParseGlob("templates/tags/components/*"))

	return &Tags{
		navbarService: navbarService,
		tagsService:   tagsService,
		log:           log,

		tagsTemplate: tagsTemplate,
	}
}

func (t *Tags) Mount(group *echo.Group) {
	group.Use(auth.RequiredMiddleware)

	group.GET("", t.tags)
	group.POST("", t.createTag)
	group.PUT("", t.updateTag)
	group.POST("/delete", t.deleteTag)

	group.RouteNotFound("/*", func(c echo.Context) error {
		// TODO: Better not found
		return c.NoContent(http.StatusNotFound)
	})
}

func (t *Tags) tags(c echo.Context) error {
	user, _ := c.Get(models.UserContextKey).(*models.User)

	navbarCtx, err := createNavbarContext(c, t.navbarService)
	if err != nil {
		t.log.Error("Failed to create navbar context", "error", err)
		return err
	}

	tags, err := t.tagsService.List(c.Request().Context(), navbarCtx.SelectedWalletId, user)
	if err != nil {
		t.log.Error("Failed to list tags", "error", err)

		// TODO: Better error handling
		return err
	}

	ctx := &models.TagsContext{
		Navbar: navbarCtx,
		Tags:   tags,
	}

	return renderTemplate(c, t.tagsTemplate, ctx)
}

func (t *Tags) createTag(c echo.Context) error {
	user, _ := c.Get(models.UserContextKey).(*models.User)

	walletId := getWalletId(c)
	name := c.FormValue("name")

	_, err := t.tagsService.Create(c.Request().Context(), walletId, name, user)
	if err == echo.ErrForbidden {
		return c.Redirect(http.StatusSeeOther, "/")
	} else if err != nil {
		t.log.Error("Failed to create tag", "error", err)

		// TODO: Better error handling
		return err
	}

	c.Response().Header().Set(HtmxHeaderTriggerAfterSettle, HtmxEventCreateSuccess)
	return t.tags(c)
}

type updateTagForm struct {
	Id   int    `form:"id"`
	Name string `form:"name"`
}

func (t *Tags) updateTag(c echo.Context) error {
	user, _ := c.Get(models.UserContextKey).(*models.User)

	var req updateTagForm
	if err := c.Bind(&req); err != nil {
		return err
	}

	_, err := t.tagsService.Update(c.Request().Context(), req.Id, req.Name, user)
	if err == echo.ErrForbidden {
		return c.Redirect(http.StatusSeeOther, "/")
	} else if err != nil {
		t.log.Error("Failed to update tag", "error", err)

		// TODO: Better error handling
		return err
	}

	c.Response().Header().Set(HtmxHeaderTriggerAfterSettle, HtmxEventUpdateSuccess)
	return t.tags(c)
}

type deleteTagForm struct {
	Id int `form:"id"`
}

func (t *Tags) deleteTag(c echo.Context) error {
	user, _ := c.Get(models.UserContextKey).(*models.User)

	var req deleteTagForm
	if err := c.Bind(&req); err != nil {
		return err
	}

	err := t.tagsService.Delete(c.Request().Context(), req.Id, user)
	if err == echo.ErrForbidden {
		return c.Redirect(http.StatusSeeOther, "/")
	} else if err != nil {
		t.log.Error("Failed to delete tag", "error", err)

		// TODO: Better error handling
		return err
	}

	c.Response().Header().Set(HtmxHeaderTriggerAfterSettle, HtmxEventDeleteSuccess)
	return t.tags(c)
}
