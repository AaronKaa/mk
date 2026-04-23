package config

import "strings"

func (cmd Command) CommandList() []string {
	if len(cmd.Commands) > 0 {
		return append([]string(nil), cmd.Commands...)
	}
	return []string{cmd.Command}
}

func (cmd Command) DisplayCommand() string {
	if len(cmd.Commands) > 0 {
		return strings.Join(cmd.Commands, " && ")
	}
	return cmd.Command
}

func (cmd Command) IsMulti() bool {
	return len(cmd.Commands) > 0
}

func (cmd Command) HasArgTemplate() bool {
	for _, command := range cmd.CommandList() {
		if strings.Contains(command, "{{args}}") ||
			strings.Contains(command, "{{arg1}}") ||
			strings.Contains(command, "{{arg2}}") ||
			strings.Contains(command, "{{arg3}}") ||
			strings.Contains(command, "{{args_prefix ") ||
			strings.Contains(command, "{{arg1_prefix ") {
			return true
		}
	}
	return false
}
