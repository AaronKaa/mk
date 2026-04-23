package app

import "github.com/AaronKaa/mk/internal/config"

func commandEnv(cfg config.Config, cmd config.Command, baseDir string) (map[string]string, error) {
	globalFileEnv, err := config.LoadEnvFiles(baseDir, cfg.EnvFile)
	if err != nil {
		return nil, err
	}
	commandFileEnv, err := config.LoadEnvFiles(baseDir, cmd.EnvFile)
	if err != nil {
		return nil, err
	}
	return mergeEnv(globalFileEnv, cfg.Env, commandFileEnv, cmd.Env), nil
}

func mergeEnv(layers ...map[string]string) map[string]string {
	hasValues := false
	for _, layer := range layers {
		if len(layer) > 0 {
			hasValues = true
			break
		}
	}
	if !hasValues {
		return nil
	}
	env := map[string]string{}
	for _, layer := range layers {
		for key, value := range layer {
			env[key] = value
		}
	}
	return env
}
