package restapp

import (
	"context"
	"fmt"

	"github.com/go-seidon/chariot/internal/app"
	"github.com/go-seidon/chariot/internal/auth"
	"github.com/go-seidon/chariot/internal/barrel"
	"github.com/go-seidon/chariot/internal/file"
	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/chariot/internal/resthandler"
	"github.com/go-seidon/chariot/internal/storage/multipart"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/encoding/base64"
	"github.com/go-seidon/provider/hashing/bcrypt"
	"github.com/go-seidon/provider/http"
	"github.com/go-seidon/provider/identifier/ksuid"
	"github.com/go-seidon/provider/logging"
	"github.com/go-seidon/provider/serialization/json"
	"github.com/go-seidon/provider/slug/goslug"
	"github.com/go-seidon/provider/validation/govalidator"
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

		goValidator := govalidator.NewValidator()
		bcryptHasher := bcrypt.NewHasher()
		ksuidIdentifier := ksuid.NewIdentifier()
		goSlugger := goslug.NewSlugger()
		jsonSerializer := json.NewSerializer()
		base64Encoder := base64.NewEncoder()
		httpClient := http.NewClient()
		clock := datetime.NewClock()

		storageRouter, err := app.NewDefaultStorageRouter(app.StorageRouterParam{
			Config:     p.Config,
			Serializer: jsonSerializer,
			Encoder:    base64Encoder,
			HttpClient: httpClient,
			Clock:      clock,
		})
		if err != nil {
			return nil, err
		}

		basicHandler := resthandler.NewBasic(resthandler.BasicParam{
			Config: &resthandler.BasicConfig{
				AppName:    config.AppName,
				AppVersion: config.AppVersion,
			},
		})
		basicGroup := echo.Group("")
		basicGroup.GET("/", basicHandler.GetAppInfo)

		authClient := auth.NewAuthClient(auth.AuthClientParam{
			Validator:  goValidator,
			Hasher:     bcryptHasher,
			Identifier: ksuidIdentifier,
			Clock:      clock,
			AuthRepo:   repo.GetAuth(),
		})
		authHandler := resthandler.NewAuth(resthandler.AuthParam{
			AuthClient: authClient,
		})
		authClientGroup := echo.Group("/v1/auth-client")
		authClientGroup.POST("", authHandler.CreateClient)
		authClientGroup.POST("/search", authHandler.SearchClient)
		authClientGroup.GET("/:id", authHandler.GetClientById)
		authClientGroup.PUT("/:id", authHandler.UpdateClientById)

		barrelClient := barrel.NewBarrel(barrel.BarrelParam{
			Validator:  goValidator,
			Identifier: ksuidIdentifier,
			Clock:      clock,
			BarrelRepo: repo.GetBarrel(),
		})
		barrelHandler := resthandler.NewBarrel(resthandler.BarrelParam{
			Barrel: barrelClient,
		})
		barrelGroup := echo.Group("/v1/barrel")
		barrelGroup.POST("", barrelHandler.CreateBarrel)
		barrelGroup.POST("/search", barrelHandler.SearchBarrel)
		barrelGroup.GET("/:id", barrelHandler.GetBarrelById)
		barrelGroup.PUT("/:id", barrelHandler.UpdateBarrelById)

		fileClient := file.NewFile(file.FileParam{
			Validator:  goValidator,
			Identifier: ksuidIdentifier,
			Clock:      clock,
			Slugger:    goSlugger,
			Router:     storageRouter,
			BarrelRepo: repo.GetBarrel(),
			FileRepo:   repo.GetFile(),
		})
		fileHandler := resthandler.NewFile(resthandler.FileParam{
			File:       fileClient,
			Serializer: jsonSerializer,
			FileParser: multipart.FileParser,
		})
		fileAccessGroup := echo.Group("/file")
		fileAccessGroup.POST("", fileHandler.UploadFile)
		fileAccessGroup.GET("/:slug", fileHandler.RetrieveFileBySlug)
	}

	app := &restApp{
		server:     server,
		config:     config,
		repository: repo,
		logger:     logger,
	}
	return app, nil
}
