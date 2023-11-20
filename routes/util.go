package routes

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func renderTemplate(c echo.Context, tmpl *template.Template, data any) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)
	return tmpl.Execute(c.Response().Writer, data)
}

func renderTemplateNamed(c echo.Context, tmpl *template.Template, name string, data any) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)
	return tmpl.ExecuteTemplate(c.Response().Writer, name, data)
}

func getWalletId(c echo.Context) int {
	walletId, _ := strconv.Atoi(c.Param("walletId"))
	return walletId
}
