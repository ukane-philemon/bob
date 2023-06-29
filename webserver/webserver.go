package webserver

import (
	"context"
	"errors"
	"fmt"
	glog "log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/ukane-philemon/bob/db"
)

// AppName is the name of the application.
const AppName = "B.O.B"

var appLog = glog.New(os.Stdout, "[webserver] ", glog.LstdFlags|glog.Lshortfile)

// Config is the configuration for the web server.
type Config struct {
	Host string `long:"host" env:"HOST" default:"127.0.0.1" description:"Server host"`
	Port string `long:"port" env:"PORT" default:"8080" description:"Server port"`
}

// WebServer is the main API server.
type WebServer struct {
	addr string
	ctx  context.Context
	*fiber.App

	db            db.DataStore
	authenticator *jwtAuthenticator
}

// New creates a new WebServer.
func New(ctx context.Context, cfg Config, db db.DataStore) (*WebServer, error) {
	if cfg.Host == "" || cfg.Port == "" {
		return nil, errors.New("invalid host or port")
	}

	a := fiber.New(fiber.Config{
		AppName:          AppName,
		Concurrency:      1000000,
		ErrorHandler:     errorHandler,
		ReadTimeout:      5 * time.Second,  // slow requests should not hold connections opened
		WriteTimeout:     60 * time.Second, // hung responses must die
		DisableKeepalive: true,
		//StrictRouting:    true,
	})

	a.Use(cors.New())
	a.Use(limiter.New(limiter.Config{
		Max:                1000,
		SkipFailedRequests: true,
	}))

	authenticator, err := newJWTAuthenticator(jwtAlg)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %w", err)
	}

	s := &WebServer{
		addr:          cfg.Host + ":" + cfg.Port,
		ctx:           ctx,
		App:           a,
		db:            db,
		authenticator: authenticator,
	}

	registerRoutes(s)
	return s, nil
}

// registerRoutes registers all the routes on WebServer.
func registerRoutes(s *WebServer) {
	s.Get("/", func(c *fiber.Ctx) error {
		return c.Status(codeOk).SendString(s.Config().AppName + " is running")
	})
	s.Get("/:shortUrl", s.handleShortUrlRedirect)

	api := s.Group("/api").Use(s.validateIfLoggedIn)

	// User Endpoints
	api.Post("/login", s.handleLogin)
	api.Get("/username-exists", s.handleUsernameExists)
	api.Post("/user", s.handleCreateAccount)
	api.Get("/user", s.handleGetUser)

	// Short URL Endpoints
	api.Post("/url", s.handleCreateShortURL)
	api.Get("/url", s.handleGetAllURL)
	api.Patch("/url", s.handleURLUpdate)
	api.Get("/url/clicks", s.handleGetShortURLClicks)
	api.Get("/url/:shortUrl", s.handleGetURL)
	api.Get("/url/:shortUrl/qr", s.handleCreateURLQR)
}

// Start starts the WebServer.
func (s *WebServer) Start() error {
	return s.Listen(s.addr)
}

// Stop stops the WebServer.
func (s *WebServer) Stop() error {
	return s.Shutdown()
}
