package config

func (source Source) VisibleCommandNames() []string {
	names := make([]string, 0, len(source.Config.Commands))
	for _, name := range SortedNames(source.Config) {
		if source.Hidden[name] {
			continue
		}
		names = append(names, name)
	}
	return names
}

func (source Source) VisibleNames() []string {
	names := make([]string, 0, len(source.Config.Commands))
	for _, name := range source.VisibleCommandNames() {
		names = append(names, name)
		names = append(names, visibleAliases(source, name)...)
	}
	return sortStrings(names)
}

func (source Source) LocalCommandNames() []string {
	var names []string
	for _, name := range source.VisibleCommandNames() {
		if !source.Inherited[name] {
			names = append(names, name)
		}
	}
	return names
}

func (source Source) InheritedCommandNames() []string {
	var names []string
	for _, name := range source.VisibleCommandNames() {
		if source.Inherited[name] {
			names = append(names, name)
		}
	}
	return names
}

func visibleAliases(source Source, name string) []string {
	cmd, ok := source.Config.Commands[name]
	if !ok {
		return nil
	}
	return append([]string(nil), cmd.Aliases...)
}
