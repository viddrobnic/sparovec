package routes

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func renderTemplate(w http.ResponseWriter, tmpl *template.Template, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return tmpl.Execute(w, data)
}

func renderTemplateNamed(w http.ResponseWriter, tmpl *template.Template, name string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return tmpl.ExecuteTemplate(w, name, data)
}

func getWalletId(r *http.Request) int {
	walletId, _ := strconv.Atoi(chi.URLParam(r, "walletId"))
	return walletId
}
