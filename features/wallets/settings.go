package wallets

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/viddrobnic/sparovec/features"
	"github.com/viddrobnic/sparovec/features/auth"
	"github.com/viddrobnic/sparovec/models"
)

func (wlts *Wallets) mountSettings(router chi.Router) {
	group := chi.NewRouter()
	group.Use(auth.RequiredMiddleware)

	group.Get("/", wlts.settings)
	group.Post("/name", wlts.settingsSaveName)
	group.Post("/add-member", wlts.settingsAddMember)
	group.Post("/remove-member", wlts.settingsRemoveMember)
	group.Post("/delete", wlts.settingsDeleteWallet)

	router.Mount("/wallets/{walletId}/settings", group)
}

func (wlts *Wallets) hasPermission(ctx context.Context, w http.ResponseWriter, walletId, userId int) bool {
	hasPermission, err := wlts.repository.HasPermission(ctx, walletId, userId)
	if err != nil {
		wlts.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return false
	}

	if !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}

	return true
}

func (wlts *Wallets) settings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	walletId := features.GetWalletId(r)
	user := auth.GetUser(r)

	if !wlts.hasPermission(ctx, w, walletId, user.Id) {
		return
	}

	wallets, err := wlts.repository.ForUser(ctx, user.Id)
	if err != nil {
		wlts.log.ErrorContext(ctx, "Failed to get wallets", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	wallet, err := wlts.repository.ForId(ctx, walletId)
	if err != nil {
		wlts.log.ErrorContext(ctx, "Failed to get wallet", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	members, err := wlts.repository.Members(ctx, walletId)
	if err != nil {
		wlts.log.ErrorContext(ctx, "Failed to get members", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := settingsViewData{
		Navbar: models.Navbar{
			SelectedWalletId: walletId,
			Wallets:          wallets,
			Username:         user.Username,
			Title:            "Šparovec | Settings",
		},
		Wallet:  wallet,
		Members: sortMembers(members, user),
	}

	view := settingsView(data)
	err = view.Render(ctx, w)
	if err != nil {
		wlts.log.ErrorContext(ctx, "Failed to render settings view", "error", err)
	}
}

func (wlts *Wallets) settingsSaveName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	walletId := features.GetWalletId(r)
	user := auth.GetUser(r)

	if !wlts.hasPermission(ctx, w, walletId, user.Id) {
		return
	}

	name := r.FormValue("name")
	err := wlts.repository.SetName(ctx, walletId, name)
	if err != nil {
		wlts.log.ErrorContext(ctx, "Failed to set name", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	wlts.settings(w, r)
}

func (wlts *Wallets) settingsAddMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	walletId := features.GetWalletId(r)
	user := auth.GetUser(r)

	if !wlts.hasPermission(ctx, w, walletId, user.Id) {
		return
	}

	username := r.FormValue("username")
	creds, err := wlts.userRepository.GetByUsername(ctx, username)
	if err != nil {
		wlts.log.ErrorContext(ctx, "Failed to get user by username", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if creds == nil {
		wlts.settings(w, r)
		return
	}

	err = wlts.repository.AddMember(ctx, walletId, creds.Id)
	if err != nil {
		wlts.log.ErrorContext(ctx, "Failed to add member", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	wlts.settings(w, r)
}

func (wlts *Wallets) settingsRemoveMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	walletId := features.GetWalletId(r)
	user := auth.GetUser(r)

	if !wlts.hasPermission(ctx, w, walletId, user.Id) {
		return
	}

	id := r.FormValue("id")
	err := wlts.repository.RemoveMember(ctx, walletId, id)
	if err != nil {
		wlts.log.ErrorContext(ctx, "Failed to remove member", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	wlts.settings(w, r)
}

func (wlts *Wallets) settingsDeleteWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	walletId := features.GetWalletId(r)
	user := auth.GetUser(r)

	if !wlts.hasPermission(ctx, w, walletId, user.Id) {
		return
	}

	err := wlts.repository.Delete(ctx, walletId)
	if err != nil {
		wlts.log.ErrorContext(ctx, "Failed to delete wallet", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func sortMembers(members []*models.Member, user *models.User) []*models.Member {
	sortedMembers := make([]*models.Member, 1, len(members))
	for _, member := range members {
		if member.Id == user.Id {
			member.IsSelf = true
			sortedMembers[0] = member
		} else {
			sortedMembers = append(sortedMembers, member)
		}
	}

	return sortedMembers
}
