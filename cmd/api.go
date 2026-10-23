package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lungtav/golang_ecommerce_backend/internals/auth"
	"github.com/lungtav/golang_ecommerce_backend/internals/db"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/mailer"
	mw "github.com/lungtav/golang_ecommerce_backend/internals/platform/middleware"
	"github.com/lungtav/golang_ecommerce_backend/internals/products"
	"github.com/lungtav/golang_ecommerce_backend/internals/stores"
)

func (app *application) mount() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	queries := db.New(app.db)
	mailer := mailer.New(app.config.resendApiKey)
	jwtSecret := app.config.jwtSecretKey
	frontendUrl := app.config.frontendUrl
	pool := app.db

	//auth
	authService := auth.NewService(queries, mailer, jwtSecret, frontendUrl, pool)
	authHandler := auth.NewHandler(authService)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/sign-up", authHandler.SignUp)
		r.Post("/sign-in", authHandler.SignIn)
		r.Post("/resend-verification", authHandler.ResendVerification)
		r.Post("/verify-email", authHandler.VerifyEmail)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
	})


	//stores
	storesService:= stores.NewService(queries)
	storesHandler:= stores.NewHandler(storesService)
	r.Route("/stores", func (r chi.Router) {
		r.Use(mw.Auth(jwtSecret))
		r.Post("/", storesHandler.CreateStore)
		r.Put("/{storeId}", storesHandler.UpdateStore)
		r.Get("/{storeId}", storesHandler.GetStore)
		r.Get("/{storeId}/products", storesHandler.GetProductsByStore)
	})

	//products
	productsService:= products.NewService(queries, pool)
	productsHandler:= products.NewHandler(productsService)
	r.Route("/products", func (r chi.Router) {
		r.Use(mw.Auth(jwtSecret))
		r.Post("/", productsHandler.CreateProduct)
	})

	return r
}

func (app *application) run(r http.Handler) error {
	srv := http.Server{
		Addr:         app.config.addr,
		Handler:      r,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Minute,
	}

	slog.Info("server is running on port", "addr", app.config.addr)

	return srv.ListenAndServe()
}
