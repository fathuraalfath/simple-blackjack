package handler

import (
	"net/http"
	"net/url"
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

	// vercel rewrites replace r.URL.Path with "/api/index.go", restore real path
	if matchedPath := r.Header.Get("x-matched-path"); matchedPath != "" {
		if u, err := url.Parse(matchedPath); err == nil {
			r.URL.Path = u.Path
			r.RequestURI = matchedPath
			if u.RawQuery != "" {
				r.URL.RawQuery = u.RawQuery
			}
		} else {
			r.URL.Path = matchedPath
			r.RequestURI = matchedPath
		}
	}

	app.ServeHTTP(w, r)
}

