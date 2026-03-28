package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	oapi "myapp/generated/server"
	"myapp/internal/db"
	"myapp/internal/deps"
	"myapp/internal/handler"
	"myapp/internal/middleware"
)

func main() {
	_ = godotenv.Load(".env.local")

	sqlDB, err := db.New()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer sqlDB.Close()

	d := deps.Deps{DB: sqlDB}

	router := chi.NewRouter()
	router.Use(cors.Handler(corsOptions()))
	router.Use(chi_middleware.Recoverer)
	router.Use(chi_middleware.Logger)

	apiRouter := chi.NewRouter()
	apiRouter.Use(middleware.Auth)
	apiRouter.Use(middleware.RateLimit)

	h := oapi.NewStrictHandlerWithOptions(
		handler.NewHandler(d),
		[]oapi.StrictMiddlewareFunc{},
		oapi.StrictHTTPServerOptions{},
	)
	oapi.HandlerFromMux(h, apiRouter)
	router.Mount("/api", apiRouter)

	// SPA fallback — serves frontend/dist for all non-API routes
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			http.NotFound(w, r)
			return
		}
		path := "./frontend/dist" + r.URL.Path
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, "./frontend/dist/index.html")
			return
		}
		http.ServeFile(w, r, path)
	})

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func corsOptions() cors.Options {
	env := os.Getenv("APP_ENV")
	if env == "production" {
		return cors.Options{
			AllowedOrigins:   []string{"https://yourdomain.com"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
			AllowCredentials: true,
			MaxAge:           600,
		}
	}
	return cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}
}
