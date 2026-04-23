package config

func NormalizeAliases(cfg *Config) error {
	if cfg.Commands == nil {
		return nil
	}
	for name, cmd := range cfg.Commands {
		if cmd.Alias == "" {
			continue
		}
		target, ok := cfg.Commands[cmd.Alias]
		if !ok {
			return Validate(*cfg)
		}
		target.Aliases = appendUnique(target.Aliases, name)
		cfg.Commands[cmd.Alias] = target
		delete(cfg.Commands, name)
	}
	return Validate(*cfg)
}

func ResolveCommand(cfg Config, name string) (string, Command, bool) {
	if cmd, ok := cfg.Commands[name]; ok {
		return name, cmd, true
	}
	for commandName, cmd := range cfg.Commands {
		for _, alias := range cmd.Aliases {
			if alias == name {
				return commandName, cmd, true
			}
		}
	}
	return "", Command{}, false
}

func AllNames(cfg Config) []string {
	names := SortedNames(cfg)
	for _, commandName := range SortedNames(cfg) {
		names = append(names, cfg.Commands[commandName].Aliases...)
	}
	return sortStrings(names)
}

func appendUnique(values []string, next string) []string {
	for _, value := range values {
		if value == next {
			return values
		}
	}
	return append(values, next)
}
