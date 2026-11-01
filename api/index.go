package handler

import (
	"net/http"
	"net/url"
	"strings"
	"sync"

	"simple-blackjack/config"
	"simple-blackjack/routers"
)

var (
	app  http.Handler
	once sync.Once
)

// entrypoint for vercel serverless
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		config.ConnectDatabase()
		app = routers.SetupRouter()
	})

	// restore original path forwarded by vercel rewrite
	if p := r.URL.Query().Get("__path"); p != "" {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		r.URL.Path = p
	} else if matchedPath := r.Header.Get("x-matched-path"); matchedPath != "" {
		if u, err := url.Parse(matchedPath); err == nil {
			r.URL.Path = u.Path
		} else {
			r.URL.Path = matchedPath
		}
	}

	app.ServeHTTP(w, r)
}

