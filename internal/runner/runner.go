package runner

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type Options struct {
	Args     []string
	Open     bool
	Dir      string
	Env      map[string]string
	Parallel bool
}

func Run(command string, args []string, open bool, dir string) error {
	return RunAll([]string{command}, Options{Args: args, Open: open, Dir: dir})
}

func RunAll(commands []string, opts Options) error {
	prepared := PreparedCommands(commands, opts.Args, opts.Open)
	return RunPrepared(prepared, opts)
}

func RunPrepared(commands []string, opts Options) error {
	if opts.Parallel {
		return runParallel(commands, opts)
	}
	for _, command := range commands {
		if err := runPrepared(command, opts); err != nil {
			return err
		}
	}
	return nil
}

func PreparedCommands(commands []string, args []string, open bool) []string {
	prepared := make([]string, len(commands))
	for i, command := range commands {
		prepared[i] = prepareCommand(command, args, open)
	}
	return prepared
}

func runPrepared(command string, opts Options) error {
	shell, flag := shellCommand()
	cmd := exec.Command(shell, flag, command)
	cmd.Dir = opts.Dir
	cmd.Env = commandEnv(opts.Env)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runParallel(commands []string, opts Options) error {
	errs := make(chan error, len(commands))
	var wg sync.WaitGroup
	for _, command := range commands {
		wg.Add(1)
		go func(command string) {
			defer wg.Done()
			errs <- runPrepared(command, opts)
		}(command)
	}
	wg.Wait()
	close(errs)

	var failures []string
	for err := range errs {
		if err != nil {
			failures = append(failures, err.Error())
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("parallel commands failed: %s", strings.Join(failures, "; "))
	}
	return nil
}

func prepareCommand(command string, args []string, open bool) string {
	templated := hasArgTemplate(command)
	if templated {
		command = expandArgs(command, args)
	}
	if open && len(args) > 0 && !templated {
		command = command + " " + quoteArgs(args)
	}
	return command
}

func expandArgs(command string, args []string) string {
	command = expandPrefixTemplates(command, args)
	return strings.NewReplacer(
		"{{args}}", quoteArgs(args),
		"{{arg1}}", arg(args, 0),
		"{{arg2}}", arg(args, 1),
		"{{arg3}}", arg(args, 2),
	).Replace(command)
}

func hasArgTemplate(command string) bool {
	return strings.Contains(command, "{{args}}") ||
		strings.Contains(command, "{{arg1}}") ||
		strings.Contains(command, "{{arg2}}") ||
		strings.Contains(command, "{{arg3}}") ||
		strings.Contains(command, "{{args_prefix ") ||
		strings.Contains(command, "{{arg1_prefix ")
}

func arg(args []string, index int) string {
	if index >= len(args) {
		return "''"
	}
	return shellQuote(args[index])
}

var prefixTemplatePattern = regexp.MustCompile(`\{\{(args|arg1)_prefix\s+"([^"]*)"\}\}`)

func expandPrefixTemplates(command string, args []string) string {
	return prefixTemplatePattern.ReplaceAllStringFunc(command, func(match string) string {
		parts := prefixTemplatePattern.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		switch parts[1] {
		case "args":
			if len(args) == 0 {
				return ""
			}
			return strings.TrimSpace(parts[2] + " " + quoteArgs(args))
		case "arg1":
			if len(args) == 0 {
				return ""
			}
			return strings.TrimSpace(parts[2] + " " + arg(args, 0))
		default:
			return match
		}
	})
}

func shellCommand() (string, string) {
	if runtime.GOOS == "windows" {
		return "cmd", "/C"
	}
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell, "-c"
	}
	return "/bin/sh", "-c"
}

func commandEnv(env map[string]string) []string {
	if len(env) == 0 {
		return os.Environ()
	}
	merged := map[string]string{}
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			merged[key] = value
		}
	}
	for key, value := range env {
		merged[key] = value
	}

	keys := make([]string, 0, len(merged))
	for key := range merged {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, key+"="+merged[key])
	}
	return out
}

func quoteArgs(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = shellQuote(arg)
	}
	return strings.Join(quoted, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if strings.IndexFunc(s, func(r rune) bool {
		return !(r == '_' || r == '-' || r == '.' || r == '/' || r == ':' || r == '=' || r == '+' || r == ',' || r == '@' || r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z')
	}) == -1 {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
