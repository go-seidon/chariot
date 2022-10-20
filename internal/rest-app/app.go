package rest_app

import (
	"context"
	"fmt"

	"github.com/go-seidon/chariot/internal/app"
	"github.com/go-seidon/provider/logging"
	"github.com/labstack/echo/v4"
)

type restApp struct {
	server Server
	logger logging.Logger
	config *RestAppConfig
}

func (a *restApp) Run(ctx context.Context) error {
	a.logger.Infof("Running %s:%s", a.config.GetAppName(), a.config.GetAppVersion())

	a.logger.Infof("Listening on: %s", a.config.GetAddress())

	err := a.server.Start(a.config.GetAddress())
	if err != nil {
		return err
	}
	return nil
}

func (a *restApp) Stop(ctx context.Context) error {
	a.logger.Infof("Stopping %s on: %s", a.config.GetAppName(), a.config.GetAddress())

	err := a.server.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}

func NewRestApp(opts ...RestAppOption) (*restApp, error) {
	p := RestAppParam{}
	for _, opt := range opts {
		opt(&p)
	}

	if p.Config == nil {
		return nil, fmt.Errorf("invalid config")
	}

	config := &RestAppConfig{
		AppName:        fmt.Sprintf("%s-rest", p.Config.AppName),
		AppVersion:     p.Config.AppVersion,
		AppHost:        p.Config.RESTAppHost,
		AppPort:        p.Config.RESTAppPort,
		UploadFormSize: p.Config.UploadFormSize,
	}

	var err error
	logger := p.Logger
	if logger == nil {
		logger, err = app.NewDefaultLog(p.Config, config.AppName)
		if err != nil {
			return nil, err
		}
	}

	server := p.Server
	if server == nil {
		echo := echo.New()
		server = &echoServer{echo}

		basicHandler := NewBasicHandler(BasicHandlerParam{
			Config: config,
		})
		basicGroup := echo.Group("")
		basicGroup.GET("/", basicHandler.GetAppInfo)
	}

	app := &restApp{
		server: server,
		config: config,
		logger: logger,
	}
	return app, nil
}
