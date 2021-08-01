package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/sndb/retwis/pkg/data"
)

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	app.errorLog.Output(2, fmt.Sprintf("%s\n%s", err.Error(), debug.Stack()))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) methodIs(w http.ResponseWriter, r *http.Request, method ...string) bool {
	for _, m := range method {
		if r.Method == m {
			return true
		}
	}
	w.Header().Set("Allow", strings.Join(method, ", "))
	app.clientError(w, http.StatusMethodNotAllowed)
	return false
}

func isLoggedIn(r *http.Request) bool {
	user := r.Context().Value(ctxKeyUser).(*data.User)
	if user == nil {
		return false
	}
	return true
}

func removeNewlines(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}

func secureCookie(c *http.Cookie) *http.Cookie {
	c.Path = "/"
	c.HttpOnly = true
	c.Secure = true
	c.SameSite = http.SameSiteStrictMode
	return c
}
