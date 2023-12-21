package features

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func GetWalletId(r *http.Request) int {
	walletId, _ := strconv.Atoi(chi.URLParam(r, "walletId"))
	return walletId
}
