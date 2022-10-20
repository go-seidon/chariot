package app

import (
	"fmt"

	"github.com/go-seidon/provider/logging"
)

func NewDefaultLog(config *Config, appName string) (logging.Logger, error) {
	if config == nil {
		return nil, fmt.Errorf("invalid config")
	}
	if appName == "" {
		return nil, fmt.Errorf("invalid app name")
	}

	opts := []logging.LogOption{}

	appOpt := logging.WithAppContext(appName, config.AppVersion)
	opts = append(opts, appOpt)

	if config.AppDebug {
		debugOpt := logging.EnableDebugging()
		opts = append(opts, debugOpt)
	}

	if config.AppEnv == ENV_LOCAL || config.AppEnv == ENV_TEST {
		prettyOpt := logging.EnablePrettyPrint()
		opts = append(opts, prettyOpt)
	}

	skipApp := logging.AddStackSkip("github.com/go-seidon/chariot/internal/app")
	skipLog := logging.AddStackSkip("github.com/go-seidon/provider/logging")
	opts = append(opts, skipApp)
	opts = append(opts, skipLog)

	return logging.NewLogrusLog(opts...), nil
}
