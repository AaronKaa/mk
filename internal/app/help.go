package app

import (
	"fmt"
	"io"
	"strings"

	"github.com/AaronKaa/mk/internal/config"
	"github.com/AaronKaa/mk/internal/ui"
)

func commandHelp(args []string, stdout io.Writer) error {
	source, err := config.FindAndLoadSource(".")
	if err != nil {
		return err
	}
	if len(args) == 0 {
		ui.PrintGlobalHelp(stdout, source)
		return nil
	}

	name, cmd, ok := config.ResolveCommand(source.Config, args[0])
	if !ok {
		return unknownCommand(args[0], source)
	}
	ui.PrintCommandHelp(stdout, name, cmd)
	return nil
}

func unknownCommand(name string, source config.Source) error {
	names := source.VisibleNames()
	if len(names) == 0 {
		return fmt.Errorf("unknown command %q; no commands are defined in mk.json or mk.yaml", name)
	}
	return fmt.Errorf("unknown command %q\n\navailable commands:\n  %s", name, strings.Join(names, "\n  "))
}
