package tags

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/viddrobnic/sparovec/features"
	"github.com/viddrobnic/sparovec/features/auth"
	"github.com/viddrobnic/sparovec/features/htmx"
	"github.com/viddrobnic/sparovec/models"
)

type WalletRepository interface {
	ForUser(ctx context.Context, userId int) ([]*models.Wallet, error)
	HasPermission(ctx context.Context, walletId, userId int) (bool, error)
}

type Repository interface {
	List(ctx context.Context, walletId int) ([]*models.Tag, error)
	Create(ctx context.Context, walletId int, name string) (*models.Tag, error)
	Get(ctx context.Context, tagId int) (*models.Tag, error)
	Update(ctx context.Context, tagId int, name string) (*models.Tag, error)
	Delete(ctx context.Context, tagId int) error
}

type Tags struct {
	walletRepository WalletRepository
	repository       Repository

	log *slog.Logger
}

func New(
	walletRepository WalletRepository,
	repository Repository,
	log *slog.Logger,
) *Tags {
	return &Tags{
		walletRepository: walletRepository,
		repository:       repository,

		log: log,
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

func (t *Tags) listTags(ctx context.Context, walletId int, user *models.User) ([]*models.Tag, error) {
	hasPermission, err := t.walletRepository.HasPermission(ctx, walletId, user.Id)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return nil, models.ErrInternalServer
	}

	if !hasPermission {
		return []*models.Tag{}, nil
	}

	return t.repository.List(ctx, walletId)
}

func (t *Tags) hasPermission(ctx context.Context, w http.ResponseWriter, walletId, userId int) bool {
	hasPermission, err := t.walletRepository.HasPermission(ctx, walletId, userId)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return false
	}

	if !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}

	return true
}

func (t *Tags) tags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(r)
	walletId := features.GetWalletId(r)

	wallets, err := t.walletRepository.ForUser(ctx, user.Id)
	if err != nil {
		t.log.Error("Failed to get wallets", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tags, err := t.listTags(ctx, walletId, user)
	if err != nil {
		t.log.Error("Failed to get tags", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	navbar := models.Navbar{
		SelectedWalletId: walletId,
		Wallets:          wallets,
		Username:         user.Username,
		Title:            "Šparovec | Tags",
	}

	view := tagsView(tags, navbar)
	err = view.Render(ctx, w)
	if err != nil {
		t.log.Error("Failed to render view", "error", err)
	}
}

func (t *Tags) createTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(r)
	walletId := features.GetWalletId(r)

	if hasPermission := t.hasPermission(ctx, w, walletId, user.Id); !hasPermission {
		return
	}

	name := r.FormValue("name")
	_, err := t.repository.Create(ctx, walletId, name)
	if err != nil {
		t.log.Error("Failed to create tag", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set(htmx.HeaderTriggerAfterSettle, htmx.EventCreateSuccess)
	t.tags(w, r)
}

func (t *Tags) updateTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(r)
	walletId := features.GetWalletId(r)

	if hasPermission := t.hasPermission(ctx, w, walletId, user.Id); !hasPermission {
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		t.log.Error("Failed to parse tag id", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	name := r.FormValue("name")

	_, err = t.repository.Update(ctx, id, name)
	if err != nil {
		t.log.Error("Failed to update tag", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set(htmx.HeaderTriggerAfterSettle, htmx.EventUpdateSuccess)
	t.tags(w, r)
}

func (t *Tags) deleteTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(r)
	walletId := features.GetWalletId(r)

	if hasPermission := t.hasPermission(ctx, w, walletId, user.Id); !hasPermission {
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		t.log.Error("Failed to parse tag id", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = t.repository.Delete(ctx, id)
	if err != nil {
		t.log.Error("Failed to delete tag", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set(htmx.HeaderTriggerAfterSettle, htmx.EventDeleteSuccess)
	t.tags(w, r)
}
