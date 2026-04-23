package ui

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AaronKaa/mk/internal/config"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle            = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	headingStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("244"))
	commandStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	inheritedCommandStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	mutedStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	usageStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

func PrintCommandList(w io.Writer, source config.Source) {
	cfg := source.Config
	path := source.Path
	title := "mk commands"
	if cfg.Header != "" {
		title = cfg.Header
	}
	fmt.Fprintln(w, titleStyle.Render(title))
	if path != "" {
		fmt.Fprintln(w, mutedStyle.Render(path))
	}
	fmt.Fprintln(w)
	printCommands(w, source)
	fmt.Fprintln(w)
	fmt.Fprintln(w, mutedStyle.Render("Run `mk --help` for usage or `mk edit` to edit commands."))
}

func PrintGlobalHelp(w io.Writer, source config.Source) {
	cfg := source.Config
	path := source.Path
	title := "mk"
	if cfg.Header != "" {
		title = cfg.Header
	}
	fmt.Fprintln(w, titleStyle.Render(title))
	fmt.Fprintln(w)
	fmt.Fprintln(w, headingStyle.Render("Usage:"))
	fmt.Fprintln(w, usageStyle.Render("  mk <command> [args...]"))
	fmt.Fprintln(w, usageStyle.Render("  mk help <command>"))
	fmt.Fprintln(w, usageStyle.Render("  mk edit"))
	fmt.Fprintln(w, usageStyle.Render("  mk init [--json|--yaml]"))
	fmt.Fprintln(w, usageStyle.Render("  mk --convert <json|yaml>"))
	fmt.Fprintln(w, usageStyle.Render("  mk --dry-run <command> [args...]"))
	fmt.Fprintln(w, usageStyle.Render("  mk completion <bash|zsh|fish>"))
	fmt.Fprintln(w, usageStyle.Render("  mk env [--json|--yaml]"))
	fmt.Fprintln(w, usageStyle.Render("  mk convert-make [Makefile] [--json|--yaml] [-o path]"))
	fmt.Fprintln(w)

	if path != "" {
		fmt.Fprintf(w, "%s %s\n\n", headingStyle.Render("Config:"), mutedStyle.Render(path))
	}

	fmt.Fprintln(w, headingStyle.Render("Commands:"))
	printCommands(w, source)
}

func printCommands(w io.Writer, source config.Source) {
	if len(source.VisibleCommandNames()) == 0 {
		fmt.Fprintln(w, mutedStyle.Render("No commands found."))
		return
	}
	printCommandSection(w, source, source.LocalCommandNames(), headingStyle, commandStyle)
	inherited := source.InheritedCommandNames()
	if len(inherited) == 0 {
		return
	}
	if len(source.LocalCommandNames()) > 0 {
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, headingStyle.Render("Inherited Commands"))
	printCommandSection(w, source, inherited, mutedStyle, inheritedCommandStyle)
}

func printCommandSection(w io.Writer, source config.Source, names []string, groupStyle lipgloss.Style, nameStyle lipgloss.Style) {
	cfg := source.Config
	sort.SliceStable(names, func(i, j int) bool {
		left := cfg.Commands[names[i]].Group
		right := cfg.Commands[names[j]].Group
		if left == "" {
			left = "Commands"
		}
		if right == "" {
			right = "Commands"
		}
		if left == right {
			return names[i] < names[j]
		}
		return left < right
	})
	lastGroup := "\x00"
	for _, name := range names {
		cmd := cfg.Commands[name]
		group := cmd.Group
		if group == "" {
			group = "Commands"
		}
		if group != lastGroup {
			if lastGroup != "\x00" {
				fmt.Fprintln(w)
			}
			if group != "Commands" {
				fmt.Fprintln(w, groupStyle.Render(group))
			}
			lastGroup = group
		}
		summary := cmd.Help
		if summary == "" {
			summary = cmd.DisplayCommand()
		}
		open := ""
		if cmd.Open {
			open = mutedStyle.Render(" [open]")
		}
		if len(cmd.Aliases) > 0 {
			open += mutedStyle.Render(" [" + joinAliases(cmd.Aliases) + "]")
		}
		fmt.Fprintf(w, "  %-18s %s%s\n", nameStyle.Render(name), mutedStyle.Render(summary), open)
	}
}

func PrintCommandHelp(w io.Writer, name string, cmd config.Command) {
	fmt.Fprintln(w, titleStyle.Render(name))
	fmt.Fprintln(w)
	if cmd.Help != "" {
		fmt.Fprintln(w, mutedStyle.Render(cmd.Help))
		fmt.Fprintln(w)
	}
	if cmd.Usage != "" {
		fmt.Fprintln(w, headingStyle.Render("Usage:"))
		fmt.Fprintln(w, usageStyle.Render("  "+cmd.Usage))
		fmt.Fprintln(w)
	}
	if cmd.IsMulti() {
		fmt.Fprintln(w, headingStyle.Render("Commands:"))
	} else {
		fmt.Fprintln(w, headingStyle.Render("Command:"))
	}
	for _, command := range cmd.CommandList() {
		fmt.Fprintf(w, "  %s\n", commandStyle.Render(command))
	}
	if cmd.Open {
		fmt.Fprintln(w)
		if cmd.IsMulti() {
			fmt.Fprintln(w, mutedStyle.Render("Extra arguments are appended to each command."))
		} else {
			fmt.Fprintln(w, mutedStyle.Render("Extra arguments are appended to this command."))
		}
	}
	printMetadata(w, cmd)
}

func printMetadata(w io.Writer, cmd config.Command) {
	if len(cmd.Aliases) == 0 && cmd.Group == "" && cmd.Dir == "" && cmd.PathForce == "" && len(cmd.Vars) == 0 && len(cmd.Env) == 0 && len(cmd.EnvFile) == 0 && len(cmd.Deps) == 0 && !cmd.Parallel && !cmd.Confirm {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, headingStyle.Render("Options:"))
	if len(cmd.Aliases) > 0 {
		fmt.Fprintf(w, "  %s %s\n", mutedStyle.Render("aliases:"), joinAliases(cmd.Aliases))
	}
	if cmd.Group != "" {
		fmt.Fprintf(w, "  %s %s\n", mutedStyle.Render("group:"), cmd.Group)
	}
	if cmd.Dir != "" {
		fmt.Fprintf(w, "  %s %s\n", mutedStyle.Render("dir:"), cmd.Dir)
	}
	if cmd.PathForce != "" {
		fmt.Fprintf(w, "  %s %s\n", mutedStyle.Render("path_force:"), cmd.PathForce)
	}
	for _, name := range config.SortedVarNames(cmd.Vars) {
		variable := cmd.Vars[name]
		value := variable.Value
		if variable.Shell != "" {
			value = "shell: " + variable.Shell
		}
		fmt.Fprintf(w, "  %s %s=%s\n", mutedStyle.Render("var:"), name, value)
	}
	if len(cmd.Deps) > 0 {
		fmt.Fprintf(w, "  %s %v\n", mutedStyle.Render("deps:"), cmd.Deps)
	}
	if cmd.Parallel {
		fmt.Fprintf(w, "  %s true\n", mutedStyle.Render("parallel:"))
	}
	if cmd.Confirm {
		fmt.Fprintf(w, "  %s true\n", mutedStyle.Render("confirm:"))
	}
	if len(cmd.Env) > 0 {
		keys := make([]string, 0, len(cmd.Env))
		for key := range cmd.Env {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(w, "  %s %s=%s\n", mutedStyle.Render("env:"), key, cmd.Env[key])
		}
	}
	for _, file := range cmd.EnvFile {
		fmt.Fprintf(w, "  %s %s\n", mutedStyle.Render("env_file:"), file)
	}
}

func joinAliases(aliases []string) string {
	out := append([]string(nil), aliases...)
	sort.Strings(out)
	return strings.Join(out, ", ")
}

func ConfigName(path string) string {
	if path == "" {
		return "mk.json"
	}
	return filepath.Base(path)
}
