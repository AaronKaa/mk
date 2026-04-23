package config

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

const (
	EnvCommandsVar       = "MK_COMMANDS"
	EnvCommandsPrefixVar = "MK_COMMANDS_PREFIX"
	EnvCommandsPath      = "MK_COMMANDS"
)

type Config struct {
	Header    string             `json:"header,omitempty" yaml:"header,omitempty"`
	Hide      bool               `json:"hide,omitempty" yaml:"hide,omitempty"`
	PathForce string             `json:"path_force,omitempty" yaml:"path_force,omitempty"`
	Vars      Vars               `json:"vars,omitempty" yaml:"vars,omitempty"`
	Env       map[string]string  `json:"env,omitempty" yaml:"env,omitempty"`
	EnvFile   EnvFiles           `json:"env_file,omitempty" yaml:"env_file,omitempty"`
	Commands  map[string]Command `json:"commands" yaml:"commands"`
}

type Source struct {
	Config    Config
	Path      string
	BaseDir   string
	Inherited map[string]bool
	Hidden    map[string]bool
	Skipped   map[string]string
}

type Command struct {
	Command   string            `json:"command,omitempty" yaml:"command,omitempty"`
	Commands  []string          `json:"commands,omitempty" yaml:"commands,omitempty"`
	Alias     string            `json:"alias,omitempty" yaml:"alias,omitempty"` // Deprecated compatibility form for alias-only entries.
	Aliases   []string          `json:"aliases,omitempty" yaml:"aliases,omitempty"`
	Open      bool              `json:"open" yaml:"open"`
	Help      string            `json:"help,omitempty" yaml:"help,omitempty"`
	Usage     string            `json:"usage,omitempty" yaml:"usage,omitempty"`
	Group     string            `json:"group,omitempty" yaml:"group,omitempty"`
	Dir       string            `json:"dir,omitempty" yaml:"dir,omitempty"`
	PathForce string            `json:"path_force,omitempty" yaml:"path_force,omitempty"`
	Vars      Vars              `json:"vars,omitempty" yaml:"vars,omitempty"`
	Env       map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	EnvFile   EnvFiles          `json:"env_file,omitempty" yaml:"env_file,omitempty"`
	Deps      []string          `json:"deps,omitempty" yaml:"deps,omitempty"`
	Parallel  bool              `json:"parallel,omitempty" yaml:"parallel,omitempty"`
	Confirm   bool              `json:"confirm,omitempty" yaml:"confirm,omitempty"`
	Hide      bool              `json:"hide,omitempty" yaml:"hide,omitempty"`
}

type Vars map[string]Variable

type Variable struct {
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
	Shell string `json:"shell,omitempty" yaml:"shell,omitempty"`
}

type EnvFiles []string

func (files *EnvFiles) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*files = EnvFiles{single}
		return nil
	}
	var multiple []string
	if err := json.Unmarshal(data, &multiple); err != nil {
		return err
	}
	*files = EnvFiles(multiple)
	return nil
}

func (files *EnvFiles) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		*files = EnvFiles{value.Value}
	case yaml.SequenceNode:
		out := make([]string, len(value.Content))
		for i, node := range value.Content {
			out[i] = node.Value
		}
		*files = EnvFiles(out)
	default:
		return fmt.Errorf("env_file must be a string or list of strings")
	}
	return nil
}
