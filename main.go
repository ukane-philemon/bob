package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/ukane-philemon/bob/api"
	"github.com/ukane-philemon/bob/db/mongodb"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		interruptChannel := make(chan os.Signal, 1)
		signal.Notify(interruptChannel, os.Interrupt)

		// Listen for the initial shutdown signal.
		sig := <-interruptChannel
		fmt.Printf("Received signal (%s). Shutting down...\n", sig)

		// Cancel the main context and all contexts created from it.
		cancel()

		// Listen for any more shutdown signals and log that shutdown has already
		// been signaled.
		for {
			<-interruptChannel
			fmt.Println("Shutdown signaled. Already shutting down...")
		}
	}()

	exitWithErr := func(err error) {
		fmt.Println(err)
		os.Exit(1)
	}

	var cfg Config
	err := parseCLIConfig(&cfg)
	if err != nil {
		exitWithErr(err)
	}

	db, err := mongodb.Connect(ctx, cfg.MongoDBConnectionURL)
	if err != nil {
		exitWithErr(err)
	}
	defer db.Close()

	r, err := api.NewRouter(ctx, db)
	if err != nil {
		db.Close()
		exitWithErr(err)
	}

	var serverError error
	go func() {
		listenAddr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
		serverError = r.Listen(listenAddr)
		if serverError == http.ErrServerClosed {
			serverError = nil // it's not an error if the server dies because of a graceful shutdown
		}

		if serverError != nil {
			cancel()
			fmt.Printf("HTTP server error: %v\n", serverError)
		}
	}()

	// Start graceful shutdown of the server on shutdown signal.
	<-ctx.Done()

	fmt.Println("Gracefully shutting down web server...")
	if err := r.Shutdown(); err != nil {
		fmt.Printf("HTTP server Shutdown error: %v\n", err)
	}
}
