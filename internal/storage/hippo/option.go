package hippo

import (
	"github.com/go-seidon/chariot/internal/storage/multipart"
	"github.com/go-seidon/provider/encoding"
	"github.com/go-seidon/provider/http"
	"github.com/go-seidon/provider/serialization"
)

type StorageOption = func(*StorageParam)

type StorageParam struct {
	Auth       *StorageAuth
	Config     *StorageConfig
	Serializer serialization.Serializer
	Encoder    encoding.Encoder
	HttpClient http.Client
	FileWriter multipart.Writer
}

type StorageAuth struct {
	ClientId     string
	ClientSecret string
}

func WithAuth(a *StorageAuth) StorageOption {
	return func(p *StorageParam) {
		p.Auth = a
	}
}

type StorageConfig struct {
	Host string
}

func WithConfig(cfg *StorageConfig) StorageOption {
	return func(p *StorageParam) {
		p.Config = cfg
	}
}

func WithSerializer(s serialization.Serializer) StorageOption {
	return func(p *StorageParam) {
		p.Serializer = s
	}
}

func WithEncoder(e encoding.Encoder) StorageOption {
	return func(p *StorageParam) {
		p.Encoder = e
	}
}

func WithHttpClient(c http.Client) StorageOption {
	return func(p *StorageParam) {
		p.HttpClient = c
	}
}

func WithWriter(fw multipart.Writer) StorageOption {
	return func(p *StorageParam) {
		p.FileWriter = fw
	}
}
