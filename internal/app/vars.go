package app

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/AaronKaa/mk/internal/config"
)

func commandVars(cfg config.Config, cmd config.Command, dir string, env map[string]string) (map[string]string, error) {
	resolved := map[string]string{}
	if err := resolveVars(resolved, cfg.Vars, dir, env); err != nil {
		return nil, err
	}
	if err := resolveVars(resolved, cmd.Vars, dir, env); err != nil {
		return nil, err
	}
	if len(resolved) == 0 {
		return nil, nil
	}
	return resolved, nil
}

func resolveVars(resolved map[string]string, vars config.Vars, dir string, env map[string]string) error {
	for _, name := range config.SortedVarNames(vars) {
		variable := vars[name]
		if variable.Shell != "" {
			value, err := runVarShell(variable.Shell, dir, mergeEnv(env, resolved))
			if err != nil {
				return fmt.Errorf("variable %q: %w", name, err)
			}
			resolved[name] = value
			continue
		}
		resolved[name] = config.ExpandVars(variable.Value, resolved)
	}
	return nil
}

func expandCommandVars(commands []string, vars map[string]string) []string {
	if len(vars) == 0 {
		return commands
	}
	out := make([]string, len(commands))
	for i, command := range commands {
		out[i] = config.ExpandVars(command, vars)
	}
	return out
}

func runVarShell(command string, dir string, env map[string]string) (string, error) {
	shell, flag := shellCommand()
	cmd := exec.Command(shell, flag, command)
	cmd.Dir = dir
	cmd.Env = envList(env)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message != "" {
			return "", fmt.Errorf("%w: %s", err, message)
		}
		return "", err
	}
	return strings.TrimRight(stdout.String(), "\r\n"), nil
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

func envList(env map[string]string) []string {
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
	out := make([]string, 0, len(merged))
	for _, key := range config.SortedStringMapKeys(merged) {
		out = append(out, key+"="+merged[key])
	}
	return out
}
