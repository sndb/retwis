package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sndb/retwis/pkg/data"
)

type ctxKey int

const (
	ctxKeyUser ctxKey = iota
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("panic recovered by middleware: %v", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Println(r.RemoteAddr, "-", r.Method, r.URL.RequestURI(), r.Proto)
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth, err := r.Cookie("auth")
		var u *data.User
		if !errors.Is(err, http.ErrNoCookie) {
			u, err = app.data.GetUserBySecret(auth.Value)
			switch {
			case errors.Is(err, data.ErrUserNotExists):
				http.SetCookie(w, secureCookie(&http.Cookie{
					Name:   "auth",
					MaxAge: -1,
				}))
				app.clientError(w, http.StatusUnauthorized)
				return
			case err != nil:
				app.serverError(w, err)
				return
			}
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, u)))
	})
}

func (app *application) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLoggedIn(r) {
			app.clientError(w, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
