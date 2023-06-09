package main

import (
	"os"

	"github.com/jessevdk/go-flags"
)

type Config struct {
	Host                 string `long:"host" env:"HOST" default:"" description:"Server host"`
	Port                 string `long:"port" env:"PORT" default:"8080" description:"Server port"`
	MongoDBConnectionURL string `short:"mongodbURL" long:"mongodbConnectionURL" env:"MONGODB_CONNECTION_URL" required:"true" description:"MongoDB connection URL"`
}

// parseCLIConfig parses the command-line arguments into the provided struct
// with go-flags tags. If the --help flag has been passed, the struct is
// described back to the terminal and the program exits using os.Exit.
func parseCLIConfig(cfg *Config) error {
	preParser := flags.NewParser(cfg, flags.HelpFlag|flags.PassDoubleDash)
	_, flagerr := preParser.Parse()

	if flagerr != nil {
		e, ok := flagerr.(*flags.Error)
		if !ok || e.Type != flags.ErrHelp {
			preParser.WriteHelp(os.Stderr)
		}
		if ok && e.Type == flags.ErrHelp {
			preParser.WriteHelp(os.Stdout)
			os.Exit(0)
		}
		return flagerr
	}
	return nil
}
