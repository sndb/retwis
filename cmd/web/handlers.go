package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sndb/retwis/pkg/data"
	"github.com/sndb/retwis/pkg/forms"
)

func (app *application) ping(w http.ResponseWriter, r *http.Request) {
	if !app.methodIs(w, r, http.MethodGet) {
		return
	}
	fmt.Fprint(w, http.StatusText(http.StatusOK))
}

func (app *application) signup(w http.ResponseWriter, r *http.Request) {
	if !app.methodIs(w, r, http.MethodPost) {
		return
	}
	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	f := forms.New(r.PostForm)
	f.Required("username")
	f.Required("password")
	if !f.Valid() {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	_, err := app.data.CreateUser(f.Get("username"), f.Get("password"))
	if err != nil {
		if errors.Is(err, data.ErrUserAlreadyExists) {
			app.clientError(w, http.StatusConflict)
		} else {
			app.serverError(w, err)
		}
		return
	}

	http.Redirect(w, r, "/login", http.StatusCreated)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	if !app.methodIs(w, r, http.MethodPost) {
		return
	}
	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	f := forms.New(r.PostForm)
	f.Required("username")
	f.Required("password")
	if !f.Valid() {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	secret, err := app.data.GetUserSecret(f.Get("username"), f.Get("password"))
	if err != nil {
		if errors.Is(err, data.ErrInvalidCredentials) {
			app.clientError(w, http.StatusForbidden)
		} else {
			app.serverError(w, err)
		}
		return
	}
	http.SetCookie(w, secureCookie(&http.Cookie{
		Name:    "auth",
		Value:   secret,
		Expires: time.Now().Add(30 * 24 * time.Hour),
	}))

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	if !app.methodIs(w, r, http.MethodPost) {
		return
	}

	http.SetCookie(w, secureCookie(&http.Cookie{
		Name:   "auth",
		MaxAge: -1,
	}))

	user := r.Context().Value(ctxKeyUser).(*data.User)
	if err := app.data.ChangeUserSecret(user.ID); err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) timeline(w http.ResponseWriter, r *http.Request) {
	if !app.methodIs(w, r, http.MethodGet) {
		return
	}

	posts, err := app.data.GetTimeline()
	if err != nil {
		app.serverError(w, err)
		return
	}

	json.NewEncoder(w).Encode(posts)
}

func (app *application) viewProfile(w http.ResponseWriter, r *http.Request) {
	if !app.methodIs(w, r, http.MethodGet) {
		return
	}

	id, err := strconv.Atoi(strings.SplitN(r.URL.Path, "/", 3)[2])
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// TODO use query parameters for paging
	ps, err := app.data.GetPosts(id, 0, 10)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if len(ps) == 0 {
		app.notFound(w)
		return
	}

	json.NewEncoder(w).Encode(ps)
}

func (app *application) createPost(w http.ResponseWriter, r *http.Request) {
	if !app.methodIs(w, r, http.MethodPost) {
		return
	}
	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("status")
	if !form.Valid() {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	u := r.Context().Value(ctxKeyUser).(*data.User)
	pid, err := app.data.CreatePost(u.ID, removeNewlines(form.Get("status")))
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/posts/%d", pid), http.StatusCreated)
}

func (app *application) viewPost(w http.ResponseWriter, r *http.Request) {
	if !app.methodIs(w, r, http.MethodGet) {
		return
	}

	id, err := strconv.Atoi(strings.SplitN(r.URL.Path, "/", 3)[2])
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	p, err := app.data.GetPost(id)
	if err != nil {
		if errors.Is(err, data.ErrPostNotExists) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	json.NewEncoder(w).Encode(p)
}
