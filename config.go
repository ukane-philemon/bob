package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/ukane-philemon/bob/db/mongodb"
	"github.com/ukane-philemon/bob/webserver"
)

type Config struct {
	WebServerCfg webserver.Config `group:"Web server" namespace:"webserver"`
	MongoDBCfg   mongodb.Config   `group:"MongoDB" namespace:"mongodb"`
	DevMode      bool             `long:"dev" env:"DEV_MODE" description:"Enable development mode"`
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
