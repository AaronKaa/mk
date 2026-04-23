package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/AaronKaa/mk/internal/config"
	"github.com/AaronKaa/mk/internal/makeconv"
	"gopkg.in/yaml.v3"
)

func convertMakefile(args []string, stdout io.Writer) error {
	opts, err := parseMakefileArgs(args)
	if err != nil {
		return err
	}

	f, err := os.Open(opts.path)
	if err != nil {
		return err
	}
	defer f.Close()

	cfg, warnings, err := makeconv.Convert(f)
	if err != nil {
		return err
	}
	if err := config.Validate(cfg); err != nil {
		return err
	}

	data, err := marshalConfig(cfg, opts.format)
	if err != nil {
		return err
	}
	if opts.output != "" {
		if _, err := os.Stat(opts.output); err == nil {
			return fmt.Errorf("%s already exists", opts.output)
		} else if !os.IsNotExist(err) {
			return err
		}
		if err := os.WriteFile(opts.output, data, 0o644); err != nil {
			return err
		}
		for _, warning := range warnings {
			fmt.Fprintf(stdout, "warning: %s\n", warning)
		}
		fmt.Fprintf(stdout, "created %s\n", opts.output)
		return nil
	}

	for _, warning := range warnings {
		fmt.Fprintf(stdout, "# warning: %s\n", warning)
	}
	_, err = stdout.Write(data)
	return err
}

type makefileOptions struct {
	path   string
	format string
	output string
}

func parseMakefileArgs(args []string) (makefileOptions, error) {
	opts := makefileOptions{path: "Makefile", format: "yaml"}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json", "-j", "json":
			opts.format = "json"
		case "--yaml", "-y", "yaml":
			opts.format = "yaml"
		case "-o", "--output":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a path", args[i])
			}
			i++
			opts.output = args[i]
		default:
			if opts.path != "Makefile" {
				return opts, fmt.Errorf("unexpected argument %q", args[i])
			}
			opts.path = args[i]
		}
	}
	return opts, nil
}

func marshalConfig(cfg config.Config, format string) ([]byte, error) {
	switch format {
	case "json":
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return nil, err
		}
		return append(data, '\n'), nil
	default:
		return yaml.Marshal(cfg)
	}
}
