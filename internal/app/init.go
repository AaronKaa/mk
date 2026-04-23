package app

import (
	"fmt"
	"io"
	"os"

	"github.com/AaronKaa/mk/internal/config"
)

func initConfig(args []string, stdout io.Writer) error {
	format := "json"
	if len(args) > 0 {
		switch args[0] {
		case "--yaml", "-y", "yaml":
			format = "yaml"
		case "--json", "-j", "json":
			format = "json"
		default:
			return fmt.Errorf("unknown init option %q; use --json or --yaml", args[0])
		}
	}

	name := "mk.json"
	if format == "yaml" {
		name = "mk.yaml"
	}
	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("%s already exists", name)
	}

	cfg := config.Config{
		Commands: map[string]config.Command{
			"test": {
				Command: "go test",
				Open:    true,
				Help:    "Run tests.",
				Usage:   "mk test [packages...]",
				Group:   "Quality",
			},
			"build": {
				Command: "go build ./...",
				Open:    false,
				Help:    "Build the project.",
				Usage:   "mk build",
				Group:   "Quality",
			},
		},
	}
	if err := config.Save(name, cfg); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "created %s\n", name)
	return nil
}
