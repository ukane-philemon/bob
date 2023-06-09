package api

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/ukane-philemon/bob/db"
)

// Router is the router for the API.
type Router struct {
	ctx context.Context
	*fiber.App
	urlDB db.DataStore
}

func NewRouter(ctx context.Context, urlDB db.DataStore) (*Router, error) {
	a := fiber.New()

	a.Use(ra)
	r := &Router{
		ctx,
		fiber.New(),
		urlDB,
	}

	return r, nil
}

// registerRoutes registers all the routes for the API.
func registerRoutes(r *Router) {
	// Users
	r.Post("/create-account", r.handleCreateAccount)
	r.Post("/login", r.handleLogin)
	r.Get("/logout", r.handleLogout)

	// Links
	r.Get("/links", r.handleGetAllLinks)
	r.Post("/links", r.handleCreateLink)
	r.Get("/links/:shortUrl", r.handleGetLink)
	r.Get("/links/:shortUrl/qr", r.handleCreateLinkQR)
}

func (r *Router) Handler() http.Handler {
	return nil
}
