package app

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDebugConfigShowsSkippedInheritedAndHiddenCommands(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("mk.json", []byte(`{"commands":{"test":{"command":"go test ./...","aliases":["t"]},"shell":{"command":"bash"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MK_COMMANDS", `{"hide":true,"commands":{"test":{"command":"npm test"},"report":{"command":"echo report","aliases":["t"]},"ok":{"command":"echo ok"},"shell":{"command":"sh"}}}`)

	var out bytes.Buffer
	if err := Run([]string{"--debug-config"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{
		"mk debug config",
		"local:",
		"  - shell",
		"  - test",
		"inherited:",
		"  - ok",
		"hidden:",
		"  - ok",
		"skipped:",
		"  - report: alias conflict",
		"  - shell: command name conflict",
		"  - test: command name conflict",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in output:\n%s", want, text)
		}
	}
}
