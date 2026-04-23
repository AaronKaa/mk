package app

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/AaronKaa/mk/internal/config"
	"gopkg.in/yaml.v3"
)

func envExport(args []string, stdout io.Writer) error {
	format := "json"
	if len(args) > 1 {
		return fmt.Errorf("usage: mk env [--json|--yaml]")
	}
	if len(args) == 1 {
		switch args[0] {
		case "--json", "-j", "json":
			format = "json"
		case "--yaml", "-y", "yaml":
			format = "yaml"
		default:
			return fmt.Errorf("unknown env option %q; use --json or --yaml", args[0])
		}
	}

	path, err := config.Find(".")
	if err != nil {
		return err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	var data []byte
	switch format {
	case "yaml":
		data, err = yaml.Marshal(cfg)
	default:
		data, err = json.Marshal(cfg)
	}
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "export %s=%s\n", config.EnvCommandsVar, shellSingleQuote(string(data)))
	return nil
}
