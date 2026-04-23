package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AaronKaa/mk/internal/config"
	"github.com/AaronKaa/mk/internal/runner"
)

const currentDirPathForce = "@"

type commandPlan struct {
	Name     string
	Commands []string
	Options  runner.Options
	Confirm  bool
}

func executeCommand(source config.Source, name string, args []string, dryRun bool, stdout io.Writer, stack map[string]bool) error {
	plans, err := planCommand(source, name, args, stack)
	if err != nil {
		return err
	}
	for _, plan := range plans {
		if dryRun {
			for _, command := range plan.Commands {
				fmt.Fprintf(stdout, "%s\n", command)
			}
			continue
		}
		if plan.Confirm && !confirm(stdout, plan.Name) {
			return fmt.Errorf("cancelled %q", plan.Name)
		}
		if err := runner.RunPrepared(plan.Commands, plan.Options); err != nil {
			return err
		}
	}
	return nil
}

func planCommand(source config.Source, name string, args []string, stack map[string]bool) ([]commandPlan, error) {
	cfg := source.Config
	commandName, cmd, ok := config.ResolveCommand(cfg, name)
	if !ok {
		return nil, unknownCommand(name, source)
	}
	if stack[commandName] {
		return nil, fmt.Errorf("dependency cycle detected at %q", commandName)
	}
	stack[commandName] = true
	defer delete(stack, commandName)

	if len(args) > 0 && !cmd.Open && !cmd.HasArgTemplate() {
		return nil, fmt.Errorf("%q does not accept extra arguments; set open: true or use argument placeholders in %s to allow this", name, filepath.Base(source.Path))
	}

	var plans []commandPlan
	for _, dep := range cmd.Deps {
		depPlans, err := planCommand(source, dep, nil, stack)
		if err != nil {
			return nil, err
		}
		plans = append(plans, depPlans...)
	}

	dir := commandDir(source, cmd)
	env, err := commandEnv(cfg, cmd, source.BaseDir)
	if err != nil {
		return nil, err
	}
	vars, err := commandVars(cfg, cmd, dir, env)
	if err != nil {
		return nil, err
	}
	env = mergeEnv(env, vars)
	commands := expandCommandVars(cmd.CommandList(), vars)

	options := runner.Options{
		Args:     args,
		Open:     cmd.Open,
		Dir:      dir,
		Env:      env,
		Parallel: cmd.Parallel,
	}
	plans = append(plans, commandPlan{
		Name:     commandName,
		Commands: runner.PreparedCommands(commands, args, cmd.Open),
		Options:  options,
		Confirm:  cmd.Confirm,
	})
	return plans, nil
}

func commandDir(source config.Source, cmd config.Command) string {
	pathForce := source.Config.PathForce
	if cmd.PathForce != "" {
		if cmd.PathForce == currentDirPathForce {
			pathForce = ""
		} else {
			pathForce = cmd.PathForce
		}
	}
	if pathForce != "" {
		if filepath.IsAbs(pathForce) {
			return pathForce
		}
		return filepath.Join(source.BaseDir, pathForce)
	}

	dir := source.BaseDir
	if cmd.Dir != "" {
		dir = cmd.Dir
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(source.BaseDir, dir)
		}
	}
	return dir
}

func confirm(stdout io.Writer, name string) bool {
	fmt.Fprintf(stdout, "Run %q? [y/N] ", name)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes"
}
