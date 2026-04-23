package config

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var varNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func (v *Variable) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err == nil {
		*v = Variable{Value: value}
		return nil
	}
	type variable Variable
	var decoded variable
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*v = Variable(decoded)
	return nil
}

func (v Variable) MarshalJSON() ([]byte, error) {
	if v.Shell == "" {
		return json.Marshal(v.Value)
	}
	type variable Variable
	return json.Marshal(variable(v))
}

func (v *Variable) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		v.Value = value.Value
		v.Shell = ""
		return nil
	case yaml.MappingNode:
		type variable Variable
		var decoded variable
		if err := value.Decode(&decoded); err != nil {
			return err
		}
		*v = Variable(decoded)
		return nil
	default:
		return fmt.Errorf("variable must be a string or mapping")
	}
}

func (v Variable) MarshalYAML() (any, error) {
	if v.Shell == "" {
		return v.Value, nil
	}
	type variable Variable
	return variable(v), nil
}

func ValidateVars(vars Vars) error {
	for name, variable := range vars {
		if !varNamePattern.MatchString(name) {
			return fmt.Errorf("variable %q must match %s", name, varNamePattern.String())
		}
		if isReservedVar(name) {
			return fmt.Errorf("variable %q is reserved for argument templates", name)
		}
		if variable.Value != "" && variable.Shell != "" {
			return fmt.Errorf("variable %q must define value or shell, not both", name)
		}
		if strings.TrimSpace(variable.Shell) == "" && variable.Shell != "" {
			return fmt.Errorf("variable %q shell command is empty", name)
		}
	}
	return nil
}

func ExpandVars(command string, vars map[string]string) string {
	for _, name := range SortedStringMapKeys(vars) {
		command = strings.ReplaceAll(command, "{{"+name+"}}", vars[name])
	}
	return command
}

func SortedVarNames(vars Vars) []string {
	return sortMapKeys(vars)
}

func SortedStringMapKeys(values map[string]string) []string {
	return sortMapKeys(values)
}

func isReservedVar(name string) bool {
	switch name {
	case "args", "arg1", "arg2", "arg3":
		return true
	default:
		return strings.HasPrefix(name, "args_prefix") || strings.HasPrefix(name, "arg1_prefix")
	}
}

func sortMapKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return sortStrings(keys)
}
