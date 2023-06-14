package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/ukane-philemon/bob/db/mongodb"
	"github.com/ukane-philemon/bob/webserver"
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

	if cfg.MongoDBConnectionURL == "" && !cfg.DevMode {
		exitWithErr(fmt.Errorf("MongoDB connection URL is required"))
	}

	db, err := mongodb.Connect(ctx, cfg.MongoDBConnectionURL)
	if err != nil {
		exitWithErr(err)
	}
	defer db.Close()

	r, err := webserver.New(ctx, cfg.WebServerCfg, db)
	if err != nil {
		db.Close()
		exitWithErr(err)
	}

	go func() {
		<-ctx.Done()
		fmt.Println("Shutting down web server...")
		if err := r.Stop(); err != nil {
			fmt.Printf("HTTP server Shutdown error: %v\n", err)
		}
	}()

	r.Start()
}
