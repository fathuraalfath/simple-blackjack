package handler

import (
	"net/http"
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
	app.ServeHTTP(w, r)
}

