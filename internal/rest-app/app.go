package rest_app

import (
	"context"
	"fmt"

	"github.com/go-seidon/chariot/internal/app"
	"github.com/go-seidon/chariot/internal/auth"
	"github.com/go-seidon/chariot/internal/repository"
	rest_handler "github.com/go-seidon/chariot/internal/rest-handler"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/hashing"
	"github.com/go-seidon/provider/identifier"
	"github.com/go-seidon/provider/logging"
	"github.com/go-seidon/provider/validation"
	"github.com/labstack/echo/v4"
)

type restApp struct {
	config     *RestAppConfig
	server     Server
	logger     logging.Logger
	repository repository.Provider
}

func (a *restApp) Run(ctx context.Context) error {
	a.logger.Infof("Running %s:%s", a.config.GetAppName(), a.config.GetAppVersion())

	a.logger.Infof("Listening on: %s", a.config.GetAddress())

	err := a.repository.Init(ctx)
	if err != nil {
		return err
	}

	err = a.server.Start(a.config.GetAddress())
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

	repo := p.Repository
	if repo == nil {
		repo, err = app.NewDefaultRepository(p.Config)
		if err != nil {
			return nil, err
		}
	}

	server := p.Server
	if server == nil {
		echo := echo.New()
		server = &echoServer{echo}

		basicHandler := rest_handler.NewBasic(rest_handler.BasicParam{
			Config: &rest_handler.BasicConfig{
				AppName:    config.AppName,
				AppVersion: config.AppVersion,
			},
		})
		basicGroup := echo.Group("")
		basicGroup.GET("/", basicHandler.GetAppInfo)

		validator := validation.NewGoValidator()
		hasher := hashing.NewBcryptHasher()
		identifier := identifier.NewKsuid()
		clock := datetime.NewClock()

		authClient := auth.NewAuthClient(auth.AuthClientParam{
			Validator:  validator,
			Hasher:     hasher,
			Identifier: identifier,
			Clock:      clock,
			AuthRepo:   repo.GetAuth(),
		})
		authHandler := rest_handler.NewAuth(rest_handler.AuthParam{
			AuthClient: authClient,
		})
		authClientGroup := echo.Group("/v1/auth-client")
		authClientGroup.POST("", authHandler.CreateClient)
		authClientGroup.GET("/:id", authHandler.GetClientById)
	}

	app := &restApp{
		server:     server,
		config:     config,
		repository: repo,
		logger:     logger,
	}
	return app, nil
}
