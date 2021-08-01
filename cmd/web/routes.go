package main

import "net/http"

func (app *application) standardMiddleware(next http.Handler) http.Handler {
	return app.recoverPanic(app.logRequest(next))
}

func (app *application) dynamicMiddleware(next http.Handler) http.Handler {
	return app.authenticate(next)
}

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/ping", app.standardMiddleware(http.HandlerFunc(app.ping)))
	mux.Handle("/signup", app.standardMiddleware(app.dynamicMiddleware(http.HandlerFunc(app.signup))))
	mux.Handle("/login", app.standardMiddleware(app.dynamicMiddleware(http.HandlerFunc(app.login))))
	mux.Handle("/logout", app.standardMiddleware(app.dynamicMiddleware(app.requireAuth(http.HandlerFunc(app.logout)))))
	mux.Handle("/timeline", app.standardMiddleware(app.dynamicMiddleware(http.HandlerFunc(app.timeline))))
	mux.Handle("/profile/", app.standardMiddleware(app.dynamicMiddleware(http.HandlerFunc(app.viewProfile))))
	mux.Handle("/posts", app.standardMiddleware(app.dynamicMiddleware(app.requireAuth(http.HandlerFunc(app.createPost)))))
	mux.Handle("/posts/", app.standardMiddleware(app.dynamicMiddleware(http.HandlerFunc(app.viewPost))))

	fs := http.StripPrefix("/static", http.FileServer(http.Dir("/static")))
	mux.Handle("/static/", fs)

	return mux
}
