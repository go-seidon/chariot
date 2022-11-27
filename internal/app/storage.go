package app

import (
	"fmt"
	"strings"

	"github.com/go-seidon/chariot/internal/barrel"
	"github.com/go-seidon/chariot/internal/storage"
	"github.com/go-seidon/chariot/internal/storage/hippo"
	"github.com/go-seidon/chariot/internal/storage/multipart"
	"github.com/go-seidon/chariot/internal/storage/router"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/encoding"
	"github.com/go-seidon/provider/http"
	"github.com/go-seidon/provider/serialization"
)

type StorageRouterParam struct {
	Config     *Config
	Serializer serialization.Serializer
	Encoder    encoding.Encoder
	HttpClient http.Client
	Clock      datetime.Clock
}

func NewDefaultStorageRouter(p StorageRouterParam) (router.Router, error) {
	if p.Config == nil {
		return nil, fmt.Errorf("invalid config")
	}

	barrels := map[string]storage.Storage{}
	r := router.NewRouter(router.RouterParam{
		Barrels: barrels,
	})
	if len(p.Config.Barrels) == 0 {
		return r, nil
	}

	for code, config := range p.Config.Barrels {
		supported := barrel.SUPPORTED_PROVIDERS[config.Provider]
		if !supported {
			return nil, fmt.Errorf("invalid barrel provider")
		}

		var stg storage.Storage
		if config.Provider == barrel.PROVIDER_HIPPO {
			stg = hippo.NewStorage(
				hippo.WithAuth(&hippo.StorageAuth{
					ClientId:     config.HippoClientId,
					ClientSecret: config.HippoClientSecret,
				}),
				hippo.WithConfig(&hippo.StorageConfig{
					Host: config.HippoServerHost,
				}),
				hippo.WithEncoder(p.Encoder),
				hippo.WithSerializer(p.Serializer),
				hippo.WithHttpClient(p.HttpClient),
				hippo.WithWriter(multipart.FileWriter),
				hippo.WithClock(p.Clock),
			)
			barrels[strings.ToLower(code)] = stg
		}
	}
	r = router.NewRouter(router.RouterParam{
		Barrels: barrels,
	})
	return r, nil
}
