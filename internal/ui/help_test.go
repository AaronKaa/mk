package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/AaronKaa/mk/internal/config"
)

func TestGlobalHelpDoesNotDuplicateCommandsHeadingForUngroupedCommands(t *testing.T) {
	t.Parallel()

	source := config.Source{
		Config: config.Config{
			Commands: map[string]config.Command{
				"test": {
					Command: "go test ./...",
					Help:    "Run tests.",
				},
			},
		},
		Path:      "mk.json",
		Inherited: map[string]bool{},
		Hidden:    map[string]bool{},
	}

	var out bytes.Buffer
	PrintGlobalHelp(&out, source)

	if got := strings.Count(out.String(), "Commands"); got != 1 {
		t.Fatalf("expected one Commands heading, got %d:\n%s", got, out.String())
	}
}

func TestGlobalHelpShowsInheritedCommandsInSeparateSectionAndHidesHiddenOnes(t *testing.T) {
	t.Parallel()

	source := config.Source{
		Config: config.Config{
			Commands: map[string]config.Command{
				"local":   {Command: "echo local", Help: "Local."},
				"ci-test": {Command: "echo inherited", Help: "Inherited."},
				"secret":  {Command: "echo hidden", Help: "Hidden."},
			},
		},
		Path: "mk.json",
		Inherited: map[string]bool{
			"ci-test": true,
			"secret":  true,
		},
		Hidden: map[string]bool{
			"secret": true,
		},
	}

	var out bytes.Buffer
	PrintGlobalHelp(&out, source)
	text := out.String()

	if !strings.Contains(text, "Inherited Commands") {
		t.Fatalf("expected inherited commands section:\n%s", text)
	}
	if !strings.Contains(text, "ci-test") {
		t.Fatalf("expected inherited command to be visible:\n%s", text)
	}
	if strings.Contains(text, "secret") {
		t.Fatalf("expected hidden inherited command to be omitted:\n%s", text)
	}
}
