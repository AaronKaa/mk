package runner

import (
	"strings"
	"testing"
)

func TestShellQuote(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"abc":          "abc",
		"foo/bar=baz":  "foo/bar=baz",
		"hello world":  "'hello world'",
		"it's working": "'it'\\''s working'",
		"":             "''",
	}

	for input, want := range tests {
		if got := shellQuote(input); got != want {
			t.Fatalf("shellQuote(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestExpandArgs(t *testing.T) {
	t.Parallel()

	got := expandArgs("run {{arg1}} then {{args}}", []string{"first file", "second"})
	want := "run 'first file' then 'first file' second"
	if got != want {
		t.Fatalf("expandArgs() = %q, want %q", got, want)
	}
}

func TestRunAllAppendsArgsPerCommand(t *testing.T) {
	t.Parallel()

	commands := []string{"echo one", "echo two {{args}}"}
	if got := prepareCommand(commands[0], []string{"filter"}, true); got != "echo one filter" {
		t.Fatalf("first command = %q", got)
	}
	if got := prepareCommand(commands[1], []string{"filter"}, true); got != "echo two filter" {
		t.Fatalf("templated command = %q", got)
	}
}

func TestPrepareCommandExpandsTemplatesWithoutOpen(t *testing.T) {
	t.Parallel()

	got := prepareCommand("git checkout main -- {{arg1}}", []string{"app/Models/User.php"}, false)
	want := "git checkout main -- app/Models/User.php"
	if got != want {
		t.Fatalf("prepareCommand() = %q, want %q", got, want)
	}
}

func TestPrepareCommandExpandsOptionalArgsPrefix(t *testing.T) {
	t.Parallel()

	command := `php artisan test {{args_prefix "--filter"}}`
	if got := prepareCommand(command, nil, false); got != "php artisan test " {
		t.Fatalf("without args = %q", got)
	}
	if got := prepareCommand(command, []string{"UserTest"}, false); got != "php artisan test --filter UserTest" {
		t.Fatalf("with args = %q", got)
	}
	if got := prepareCommand(command, []string{"User Test"}, false); got != "php artisan test --filter 'User Test'" {
		t.Fatalf("with quoted args = %q", got)
	}
}

func TestPrepareCommandExpandsOptionalArg1Prefix(t *testing.T) {
	t.Parallel()

	command := `php artisan test {{arg1_prefix "--filter"}}`
	if got := prepareCommand(command, nil, false); got != "php artisan test " {
		t.Fatalf("without args = %q", got)
	}
	if got := prepareCommand(command, []string{"UserTest", "ignored"}, false); got != "php artisan test --filter UserTest" {
		t.Fatalf("with args = %q", got)
	}
}

func TestPreparedCommands(t *testing.T) {
	t.Parallel()

	got := PreparedCommands([]string{"go test", "go vet"}, []string{"./..."}, true)
	want := []string{"go test ./...", "go vet ./..."}
	if len(got) != len(want) {
		t.Fatalf("PreparedCommands() length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("PreparedCommands()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCommandEnvOverridesExistingEnvironment(t *testing.T) {
	t.Setenv("MK_RUNNER_TEST_ENV", "parent")

	env := commandEnv(map[string]string{"MK_RUNNER_TEST_ENV": "child"})
	count := 0
	value := ""
	for _, entry := range env {
		key, got, ok := strings.Cut(entry, "=")
		if ok && key == "MK_RUNNER_TEST_ENV" {
			count++
			value = got
		}
	}
	if count != 1 {
		t.Fatalf("expected one env entry, got %d in %#v", count, env)
	}
	if value != "child" {
		t.Fatalf("env value = %q, want child", value)
	}
}
