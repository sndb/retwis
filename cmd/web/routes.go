package main

import "net/http"

func (app *application) routes() http.Handler {
	type h = http.HandlerFunc

	mux := http.NewServeMux()

	mux.Handle("/ping", h(app.ping))
	mux.Handle("/signup", app.dynamicMiddleware(h(app.signup)))
	mux.Handle("/login", app.dynamicMiddleware(h(app.login)))
	mux.Handle("/logout", app.dynamicMiddleware(app.requireAuth(h(app.logout))))
	mux.Handle("/timeline", app.dynamicMiddleware(h(app.timeline)))
	mux.Handle("/profile/", app.dynamicMiddleware(h(app.viewProfile)))
	mux.Handle("/posts", app.dynamicMiddleware(app.requireAuth(h(app.createPost))))
	mux.Handle("/posts/", app.dynamicMiddleware(h(app.viewPost)))

	fs := http.StripPrefix("/static", http.FileServer(http.Dir("/static")))
	mux.Handle("/static/", fs)

	return app.standardMiddleware(mux)
}
