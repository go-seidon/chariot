package app

import (
	"fmt"

	"github.com/go-seidon/provider/logging"
	"github.com/go-seidon/provider/logging/logrus"
)

func NewDefaultLog(config *Config, appName string) (logging.Logger, error) {
	if config == nil {
		return nil, fmt.Errorf("invalid config")
	}
	if appName == "" {
		return nil, fmt.Errorf("invalid app name")
	}

	opts := []logrus.LogOption{}

	appOpt := logrus.WithAppContext(appName, config.AppVersion)
	opts = append(opts, appOpt)

	if config.AppDebug {
		debugOpt := logrus.EnableDebugging()
		opts = append(opts, debugOpt)
	}

	if config.AppEnv == ENV_LOCAL || config.AppEnv == ENV_TEST {
		prettyOpt := logrus.EnablePrettyPrint()
		opts = append(opts, prettyOpt)
	}

	skipApp := logrus.AddStackSkip("github.com/go-seidon/chariot/internal/app")
	skipLog := logrus.AddStackSkip("github.com/go-seidon/provider/logging")
	opts = append(opts, skipApp)
	opts = append(opts, skipLog)

	return logrus.NewLogger(opts...), nil
}
