package config

type MergeResult struct {
	Inherited map[string]bool
	Hidden    map[string]bool
	Skipped   map[string]string
}

func MergePrefixed(dst *Config, src Config, prefix string) (MergeResult, error) {
	if dst.Commands == nil {
		dst.Commands = map[string]Command{}
	}
	result := MergeResult{
		Inherited: map[string]bool{},
		Hidden:    map[string]bool{},
		Skipped:   map[string]string{},
	}
	occupied := occupiedNames(*dst)
	prefixed := map[string]Command{}
	for name, cmd := range src.Commands {
		nextName := prefix + name
		if occupied[nextName] {
			result.Skipped[nextName] = "command name conflict"
			continue
		}
		aliases := append([]string(nil), cmd.Aliases...)
		conflict := false
		for i, alias := range aliases {
			aliases[i] = prefix + alias
			if occupied[aliases[i]] {
				result.Skipped[nextName] = "alias conflict"
				conflict = true
				break
			}
		}
		if conflict {
			continue
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
		cmd.Aliases = aliases
		for i, dep := range cmd.Deps {
			cmd.Deps[i] = prefix + dep
		}
		prefixed[nextName] = cmd
		result.Inherited[nextName] = true
		if src.Hide || cmd.Hide {
			result.Hidden[nextName] = true
		}
		occupied[nextName] = true
		for _, alias := range aliases {
			occupied[alias] = true
		}
	}
	for {
		changed := false
		for name, cmd := range prefixed {
			if hasMissingDeps(*dst, prefixed, cmd) {
				result.Skipped[name] = "missing dependency after merge"
				delete(result.Inherited, name)
				delete(result.Hidden, name)
				delete(prefixed, name)
				changed = true
			}
		}
		if !changed {
			break
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

func occupiedNames(cfg Config) map[string]bool {
	out := map[string]bool{}
	for name := range cfg.Commands {
		out[name] = true
	}
	for _, name := range AllNames(cfg) {
		out[name] = true
	}
	return out
}

func hasMissingDeps(dst Config, prefixed map[string]Command, cmd Command) bool {
	for _, dep := range cmd.Deps {
		if _, ok := dst.Commands[dep]; ok {
			continue
		}
		if _, ok := prefixed[dep]; ok {
			continue
		}
		return true
	}
	return false
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
