package app

import (
	"fmt"
	"io"

	"github.com/AaronKaa/mk/internal/config"
)

func debugConfig(stdout io.Writer) error {
	source, err := loadOptionalConfig(".")
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, "mk debug config")
	fmt.Fprintf(stdout, "path: %s\n", displayValue(source.Path, "(none)"))
	fmt.Fprintf(stdout, "base_dir: %s\n", displayValue(source.BaseDir, "(none)"))
	fmt.Fprintf(stdout, "commands: %d\n", len(source.Config.Commands))

	printSortedMapKeys(stdout, "local", source.LocalCommandNames())
	printSortedMapKeys(stdout, "inherited", source.InheritedCommandNames())
	printSortedMapKeys(stdout, "hidden", hiddenVisibleNames(source))
	printSkipped(stdout, source.Skipped)
	return nil
}

func printSortedMapKeys(stdout io.Writer, label string, values []string) {
	fmt.Fprintf(stdout, "%s:", label)
	if len(values) == 0 {
		fmt.Fprintln(stdout, " (none)")
		return
	}
	fmt.Fprintln(stdout)
	for _, value := range values {
		fmt.Fprintf(stdout, "  - %s\n", value)
	}
}

func printSkipped(stdout io.Writer, skipped map[string]string) {
	fmt.Fprint(stdout, "skipped:")
	if len(skipped) == 0 {
		fmt.Fprintln(stdout, " (none)")
		return
	}
	fmt.Fprintln(stdout)
	for _, name := range config.SortedStringMapKeys(skipped) {
		fmt.Fprintf(stdout, "  - %s: %s\n", name, skipped[name])
	}
}

func hiddenVisibleNames(source config.Source) []string {
	names := make([]string, 0, len(source.Hidden))
	for _, name := range config.SortedNames(source.Config) {
		if source.Hidden[name] {
			names = append(names, name)
		}
	}
	return names
}

func displayValue(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
