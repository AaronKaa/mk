package config

import (
	"errors"
	"fmt"
	"strings"
)

func Validate(cfg Config) error {
	if err := ValidateVars(cfg.Vars); err != nil {
		return err
	}
	aliases := map[string]string{}
	for name, cmd := range cfg.Commands {
		if err := cmd.Validate(); err != nil {
			return fmt.Errorf("command %q: %w", name, err)
		}
		if cmd.Alias != "" {
			if _, ok := cfg.Commands[cmd.Alias]; !ok {
				return fmt.Errorf("command %q: alias target %q does not exist", name, cmd.Alias)
			}
		}
		for _, alias := range cmd.Aliases {
			if strings.TrimSpace(alias) == "" {
				return fmt.Errorf("command %q: aliases cannot contain empty values", name)
			}
			if _, exists := cfg.Commands[alias]; exists {
				return fmt.Errorf("command %q: alias %q conflicts with a command name", name, alias)
			}
			if existing, exists := aliases[alias]; exists {
				return fmt.Errorf("command %q: alias %q already belongs to command %q", name, alias, existing)
			}
			aliases[alias] = name
		}
		for _, dep := range cmd.Deps {
			if _, ok := cfg.Commands[dep]; !ok {
				return fmt.Errorf("command %q: dependency %q does not exist", name, dep)
			}
		}
	}
	return nil
}

func (cmd Command) Validate() error {
	if cmd.Alias != "" {
		if cmd.Command != "" || len(cmd.Commands) > 0 {
			return errors.New("alias commands cannot also define command or commands")
		}
		if len(cmd.Aliases) > 0 {
			return errors.New("alias commands cannot also define aliases")
		}
		return nil
	}
	if cmd.Command != "" && len(cmd.Commands) > 0 {
		return errors.New("use command or commands, not both")
	}
	if cmd.Command == "" && len(cmd.Commands) == 0 {
		return errors.New("missing command or commands field")
	}
	if strings.TrimSpace(cmd.Command) == "" && len(cmd.Commands) == 0 {
		return errors.New("command field is empty")
	}
	for i, command := range cmd.Commands {
		if strings.TrimSpace(command) == "" {
			return fmt.Errorf("commands[%d] is empty", i)
		}
	}
	if err := ValidateVars(cmd.Vars); err != nil {
		return err
	}
	return nil
}
