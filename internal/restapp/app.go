package restapp

import (
	"context"
	"fmt"

	"github.com/go-seidon/chariot/internal/app"
	"github.com/go-seidon/chariot/internal/auth"
	"github.com/go-seidon/chariot/internal/barrel"
	"github.com/go-seidon/chariot/internal/file"
	"github.com/go-seidon/chariot/internal/queue"
	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/chariot/internal/resthandler"
	"github.com/go-seidon/chariot/internal/restmiddleware"
	"github.com/go-seidon/chariot/internal/session"
	"github.com/go-seidon/chariot/internal/signature/jwt"
	"github.com/go-seidon/chariot/internal/storage/multipart"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/encoding/base64"
	"github.com/go-seidon/provider/hashing/bcrypt"
	"github.com/go-seidon/provider/http"
	"github.com/go-seidon/provider/identifier/ksuid"
	"github.com/go-seidon/provider/logging"
	"github.com/go-seidon/provider/serialization"
	"github.com/go-seidon/provider/serialization/json"
	"github.com/go-seidon/provider/serialization/protobuf"
	"github.com/go-seidon/provider/slug/goslug"
	"github.com/go-seidon/provider/validation/govalidator"
	"github.com/labstack/echo/v4"
)

type restApp struct {
	config     *RestAppConfig
	server     Server
	logger     logging.Logger
	repository repository.Provider
	queue      queue.Queue
}

func (a *restApp) Run(ctx context.Context) error {
	a.logger.Infof("Running %s:%s", a.config.GetAppName(), a.config.GetAppVersion())

	a.logger.Infof("Initializing repository")
	err := a.repository.Init(ctx)
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
		listeners <- a.server.Start(a.config.GetAddress())
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

	var fileClient file.File
	var protobufSerializer serialization.Serializer

	server := p.Server
	if server == nil {
		e := echo.New()
		server = &echoServer{e}

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

		barrelClient := barrel.NewBarrel(barrel.BarrelParam{
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

		fileClient = file.NewFile(file.FileParam{
			Config: &file.FileConfig{
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
		basicGroup.GET("/", basicHandler.GetAppInfo)

		basicAuthGroup := e.Group("/v1", basicAuthMiddleware)
		basicAuthGroup.POST("/auth-client", authHandler.CreateClient)
		basicAuthGroup.POST("/auth-client/search", authHandler.SearchClient)
		basicAuthGroup.GET("/auth-client/:id", authHandler.GetClientById)
		basicAuthGroup.PUT("/auth-client/:id", authHandler.UpdateClientById)
		basicAuthGroup.POST("/barrel", barrelHandler.CreateBarrel)
		basicAuthGroup.POST("/barrel/search", barrelHandler.SearchBarrel)
		basicAuthGroup.GET("/barrel/:id", barrelHandler.GetBarrelById)
		basicAuthGroup.PUT("/barrel/:id", barrelHandler.UpdateBarrelById)
		basicAuthGroup.POST("/session", sessionHandler.CreateSession)
		basicAuthGroup.GET("/file/:id", fileHandler.GetFileById)
		basicAuthGroup.POST("/file/search", fileHandler.SearchFile)
		basicAuthGroup.POST("/file/replication", fileHandler.ScheduleReplication)

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
		server:     server,
		config:     config,
		repository: repo,
		logger:     logger,
		queue:      que,
	}
	return app, nil
}
