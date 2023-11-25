package routes

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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

func (t *Tags) Mount(router chi.Router) {
	group := chi.NewRouter()
	group.Use(auth.RequiredMiddleware)

	group.Get("/", t.tags)
	group.Post("/", t.createTag)
	group.Put("/", t.updateTag)
	group.Post("/delete", t.deleteTag)

	router.Mount("/wallets/{walletId}/tags", group)
}

func (t *Tags) tags(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)

	navbarCtx, err := createNavbarContext(r, t.navbarService)
	if err != nil {
		t.log.Error("Failed to create navbar context", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tags, err := t.tagsService.List(r.Context(), navbarCtx.SelectedWalletId, user)
	if err != nil {
		t.log.Error("Failed to list tags", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	ctx := &models.TagsContext{
		Navbar: navbarCtx,
		Tags:   tags,
	}

	err = renderTemplate(w, t.tagsTemplate, ctx)
	if err != nil {
		t.log.Error("Failed to render template", "error", err)
	}
}

func (t *Tags) createTag(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)

	walletId := getWalletId(r)
	name := r.FormValue("name")

	_, err := t.tagsService.Create(r.Context(), walletId, name, user)
	if err == models.ErrForbidden {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else if err != nil {
		t.log.Error("Failed to create tag", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set(HtmxHeaderTriggerAfterSettle, HtmxEventCreateSuccess)
	t.tags(w, r)
}

func (t *Tags) updateTag(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		t.log.Error("Failed to parse tag id", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	name := r.FormValue("name")

	_, err = t.tagsService.Update(r.Context(), id, name, user)
	if err == models.ErrForbidden {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else if err != nil {
		t.log.Error("Failed to update tag", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set(HtmxHeaderTriggerAfterSettle, HtmxEventUpdateSuccess)
	t.tags(w, r)
}

func (t *Tags) deleteTag(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		t.log.Error("Failed to parse tag id", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = t.tagsService.Delete(r.Context(), id, user)
	if err == models.ErrForbidden {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else if err != nil {
		t.log.Error("Failed to delete tag", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set(HtmxHeaderTriggerAfterSettle, HtmxEventDeleteSuccess)
	t.tags(w, r)
}
