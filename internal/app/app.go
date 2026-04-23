package app

import (
	"fmt"
	"io"

	"github.com/AaronKaa/mk/internal/config"
	"github.com/AaronKaa/mk/internal/editor"
	"github.com/AaronKaa/mk/internal/ui"
)

var version = "0.1.0"

func Run(args []string, stdout, stderr io.Writer) error {
	dryRun := false
	if len(args) > 0 && args[0] == "--dry-run" {
		dryRun = true
		args = args[1:]
	}

	if len(args) == 0 {
		source, err := loadOptionalConfig(".")
		if err != nil {
			return err
		}
		ui.PrintCommandList(stdout, source)
		return nil
	}

	if args[0] == "--help" || args[0] == "-h" {
		source, err := loadOptionalConfig(".")
		if err != nil {
			return err
		}
		ui.PrintGlobalHelp(stdout, source)
		return nil
	}

	if args[0] == "--convert" {
		return convertConfig(args[1:], stdout)
	}

	if args[0] == "--debug-config" {
		return debugConfig(stdout)
	}

	switch args[0] {
	case "--version", "-v", "version":
		fmt.Fprintf(stdout, "mk %s\n", version)
		return nil
	case "init":
		return initConfig(args[1:], stdout)
	case "edit":
		return editor.Run(args[1:])
	case "help":
		return commandHelp(args[1:], stdout)
	case "completion":
		return completion(args[1:], stdout)
	case "env":
		return envExport(args[1:], stdout)
	case "convert-make":
		return convertMakefile(args[1:], stdout)
	}

	source, err := config.FindAndLoadSource(".")
	if err != nil {
		return err
	}

	return executeCommand(source, args[0], args[1:], dryRun, stdout, map[string]bool{})
}

func loadOptionalConfig(start string) (config.Source, error) {
	source, err := config.FindAndLoadSource(start)
	if err == nil {
		return source, nil
	}
	if config.IsNotFound(err) {
		return config.Source{Config: config.Config{Commands: map[string]config.Command{}}, Path: "", BaseDir: ".", Inherited: map[string]bool{}, Hidden: map[string]bool{}, Skipped: map[string]string{}}, nil
	}
	return config.Source{}, err
}
