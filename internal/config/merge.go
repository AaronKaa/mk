package config

import "fmt"

type MergeResult struct {
	Inherited map[string]bool
	Hidden    map[string]bool
}

func MergePrefixed(dst *Config, src Config, prefix string) (MergeResult, error) {
	if dst.Commands == nil {
		dst.Commands = map[string]Command{}
	}
	result := MergeResult{
		Inherited: map[string]bool{},
		Hidden:    map[string]bool{},
	}
	prefixed := map[string]Command{}
	for name, cmd := range src.Commands {
		nextName := prefix + name
		if _, exists := dst.Commands[nextName]; exists {
			return MergeResult{}, fmt.Errorf("env command %q conflicts with file command", nextName)
		}
		if src.PathForce != "" && cmd.PathForce == "" {
			cmd.PathForce = src.PathForce
		}
		if len(src.Vars) > 0 {
			cmd.Vars = mergeVars(src.Vars, cmd.Vars)
		}
		if len(src.EnvFile) > 0 {
			cmd.EnvFile = append(append(EnvFiles{}, src.EnvFile...), cmd.EnvFile...)
		}
		if len(src.Env) > 0 {
			cmd.Env = mergeStringMap(src.Env, cmd.Env)
		}
		for i, alias := range cmd.Aliases {
			cmd.Aliases[i] = prefix + alias
		}
		for i, dep := range cmd.Deps {
			cmd.Deps[i] = prefix + dep
		}
		prefixed[nextName] = cmd
		result.Inherited[nextName] = true
		if src.Hide || cmd.Hide {
			result.Hidden[nextName] = true
		}
	}
	for name, cmd := range prefixed {
		dst.Commands[name] = cmd
	}
	if err := Validate(*dst); err != nil {
		return MergeResult{}, err
	}
	return result, nil
}

func mergeVars(layers ...Vars) Vars {
	var out Vars
	for _, layer := range layers {
		for key, value := range layer {
			if out == nil {
				out = Vars{}
			}
			out[key] = value
		}
	}
	return out
}

func mergeStringMap(layers ...map[string]string) map[string]string {
	var out map[string]string
	for _, layer := range layers {
		for key, value := range layer {
			if out == nil {
				out = map[string]string{}
			}
			out[key] = value
		}
	}
	return out
}
