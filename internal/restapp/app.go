package restapp

import (
	"context"
	"fmt"
	net_http "net/http"

	"github.com/go-seidon/chariot/internal/app"
	"github.com/go-seidon/chariot/internal/auth"
	"github.com/go-seidon/chariot/internal/healthcheck"
	"github.com/go-seidon/chariot/internal/queue"
	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/chariot/internal/resthandler"
	"github.com/go-seidon/chariot/internal/restmiddleware"
	"github.com/go-seidon/chariot/internal/service"
	"github.com/go-seidon/chariot/internal/session"
	"github.com/go-seidon/chariot/internal/signature/jwt"
	"github.com/go-seidon/chariot/internal/storage/multipart"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/encoding/base64"
	"github.com/go-seidon/provider/hashing/bcrypt"
	"github.com/go-seidon/provider/health"
	"github.com/go-seidon/provider/http"
	"github.com/go-seidon/provider/identity/ksuid"
	"github.com/go-seidon/provider/logging"
	"github.com/go-seidon/provider/serialization"
	"github.com/go-seidon/provider/serialization/json"
	"github.com/go-seidon/provider/serialization/protobuf"
	"github.com/go-seidon/provider/slug/goslug"
	"github.com/go-seidon/provider/validation/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/bytes"
)

type restApp struct {
	config       *RestAppConfig
	server       Server
	logger       logging.Logger
	repository   repository.Repository
	queue        queue.Queue
	healthClient health.HealthCheck
}

func (a *restApp) Run(ctx context.Context) error {
	a.logger.Infof("Running %s:%s", a.config.GetAppName(), a.config.GetAppVersion())

	a.logger.Infof("Initializing repository")
	err := a.repository.Init(ctx)
	if err != nil {
		return err
	}

	err = a.healthClient.Start(ctx)
	if err != nil {
		return err
	}

	listeners := make(chan error, 2)
	go func() {
		a.logger.Infof("Listening on queue")
		listeners <- a.queue.Start(ctx)
	}()

	go func() {
		a.logger.Infof("Listening on: %s", a.config.GetAddress())
		err := a.server.Start(a.config.GetAddress())
		if err != net_http.ErrServerClosed {
			listeners <- err
		}
	}()

	for i := 0; i < 2; i++ {
		err := <-listeners
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *restApp) Stop(ctx context.Context) error {
	a.logger.Infof("Stopping %s on: %s", a.config.GetAppName(), a.config.GetAddress())

	err := a.healthClient.Stop(ctx)
	if err != nil {
		a.logger.Errorf("Failed stopping healthcheck, err: %s", err.Error())
	}

	err = a.server.Shutdown(ctx)
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
		AppName:         fmt.Sprintf("%s-rest", p.Config.AppName),
		AppVersion:      p.Config.AppVersion,
		AppHost:         p.Config.RESTAppHost,
		AppPort:         p.Config.RESTAppPort,
		StorageFormSize: p.Config.StorageFormSize,
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

	queuer := p.Queuer
	if queuer == nil {
		queuer, err = app.NewDefaultQueueing(p.Config)
		if err != nil {
			return nil, err
		}
	}

	healthClient := p.HealthClient
	if healthClient == nil {
		healthClient, err = app.NewDefaultHealthCheck(logger, repo)
		if err != nil {
			return nil, err
		}
	}

	var fileClient service.File
	var protobufSerializer serialization.Serializer

	server := p.Server
	if server == nil {
		e := echo.New()
		e.Debug = p.Config.AppDebug
		e.HTTPErrorHandler = NewErrorHandler(ErrorHandlerParam{
			Debug:  p.Config.AppDebug,
			Logger: logger,
		})
		e.Use(middleware.Recover())
		e.Use(middleware.RequestID())
		e.Use(middleware.CORS())
		e.Use(middleware.Secure())
		e.Use(middleware.BodyLimitWithConfig(middleware.BodyLimitConfig{
			Limit: bytes.Format(p.Config.StorageFormSize + 1024), // set form max size + add 1KB non file field
		}))
		e.Use(NewRequestLog(RequestLogParam{
			Logger: logger,
		}))
		server = e

		goValidator := govalidator.NewValidator()
		bcryptHasher := bcrypt.NewHasher()
		ksuidIdentifier := ksuid.NewIdentifier()
		goSlugger := goslug.NewSlugger()
		jsonSerializer := json.NewSerializer()
		protobufSerializer = protobuf.NewSerializer()
		base64Encoder := base64.NewEncoder()
		httpClient := http.NewClient()
		clock := datetime.NewClock()
		jwtSignature := jwt.NewSignature(
			jwt.WithIssuer(p.Config.SignatureIssuer),
			jwt.WithSignKey([]byte(p.Config.SignatureKey)),
			jwt.WithClock(clock),
		)

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

		basicClient := auth.NewBasicAuth(auth.NewBasicAuthParam{
			Encoder:  base64Encoder,
			Hasher:   bcryptHasher,
			AuthRepo: repo.GetAuth(),
		})
		basicAuth := restmiddleware.NewBasicAuth(restmiddleware.BasicAuthParam{
			BasicClient: basicClient,
			Serializer:  jsonSerializer,
		})

		basicHandler := resthandler.NewBasic(resthandler.BasicParam{
			Config: &resthandler.BasicConfig{
				AppName:    config.AppName,
				AppVersion: config.AppVersion,
			},
		})
		healthCheck := healthcheck.NewHealthCheck(healthcheck.HealthCheckParam{
			HealthClient: healthClient,
		})
		healthHandler := resthandler.NewHealth(resthandler.HealthParam{
			HealthClient: healthCheck,
		})

		authClient := service.NewAuthClient(service.AuthClientParam{
			Validator:  goValidator,
			Hasher:     bcryptHasher,
			Identifier: ksuidIdentifier,
			Clock:      clock,
			AuthRepo:   repo.GetAuth(),
		})
		authHandler := resthandler.NewAuth(resthandler.AuthParam{
			AuthClient: authClient,
		})

		barrelClient := service.NewBarrel(service.BarrelParam{
			Validator:  goValidator,
			Identifier: ksuidIdentifier,
			Clock:      clock,
			BarrelRepo: repo.GetBarrel(),
		})
		barrelHandler := resthandler.NewBarrel(resthandler.BarrelParam{
			Barrel: barrelClient,
		})

		sessionClient := session.NewSession(session.SessionParam{
			Validator:  goValidator,
			Signature:  jwtSignature,
			Clock:      clock,
			Identifier: ksuidIdentifier,
		})

		fileClient = service.NewFile(service.FileParam{
			Config: &service.FileConfig{
				AppHost: p.Config.StorageAccessHost,
			},
			Validator:     goValidator,
			Identifier:    ksuidIdentifier,
			SessionClient: sessionClient,
			Clock:         clock,
			Serializer:    protobufSerializer,
			Slugger:       goSlugger,
			Router:        storageRouter,
			BarrelRepo:    repo.GetBarrel(),
			FileRepo:      repo.GetFile(),
			Pubsub:        queuer,
		})
		fileHandler := resthandler.NewFile(resthandler.FileParam{
			File:       fileClient,
			Serializer: jsonSerializer,
			FileParser: multipart.FileParser,
		})

		uploadAuth := restmiddleware.NewSessionAuth(restmiddleware.SessionAuthParam{
			SessionClient: sessionClient,
			Serializer:    jsonSerializer,
			Feature:       "upload_file",
		})
		sessionHandler := resthandler.NewSession(resthandler.SessionParam{
			Session: sessionClient,
		})

		basicAuthMiddleware := echo.WrapMiddleware(basicAuth.Handle)
		uploadMiddleware := echo.WrapMiddleware(uploadAuth.Handle)

		basicGroup := e.Group("")
		basicGroup.GET("/info", basicHandler.GetAppInfo)

		basicAuthGroup := e.Group("", basicAuthMiddleware)
		basicAuthGroup.GET("/health", healthHandler.CheckHealth)
		basicAuthGroup.POST("/v1/auth-client", authHandler.CreateClient)
		basicAuthGroup.POST("/v1/auth-client/search", authHandler.SearchClient)
		basicAuthGroup.GET("/v1/auth-client/:id", authHandler.GetClientById)
		basicAuthGroup.PUT("/v1/auth-client/:id", authHandler.UpdateClientById)
		basicAuthGroup.POST("/v1/barrel", barrelHandler.CreateBarrel)
		basicAuthGroup.POST("/v1/barrel/search", barrelHandler.SearchBarrel)
		basicAuthGroup.GET("/v1/barrel/:id", barrelHandler.GetBarrelById)
		basicAuthGroup.PUT("/v1/barrel/:id", barrelHandler.UpdateBarrelById)
		basicAuthGroup.POST("/v1/session", sessionHandler.CreateSession)
		basicAuthGroup.GET("/v1/file/:id", fileHandler.GetFileById)
		basicAuthGroup.DELETE("/v1/file/:id", fileHandler.DeleteFileById)
		basicAuthGroup.POST("/v1/file/search", fileHandler.SearchFile)
		basicAuthGroup.POST("/v1/file/replication", fileHandler.ScheduleReplication)

		e.POST("/file", fileHandler.UploadFile, uploadMiddleware)
		e.GET("/file/:slug", fileHandler.RetrieveFileBySlug)
	}

	que := p.Queue
	if que == nil {
		que = queue.NewQueue(queue.QueueParam{
			Queuer:     queuer,
			File:       fileClient,
			Serializer: protobufSerializer,
		})
	}

	app := &restApp{
		server:       server,
		config:       config,
		repository:   repo,
		logger:       logger,
		queue:        que,
		healthClient: healthClient,
	}
	return app, nil
}
