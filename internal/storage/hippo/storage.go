package hippo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	net_http "net/http"
	"time"

	"github.com/go-seidon/chariot/internal/storage"
	"github.com/go-seidon/chariot/internal/storage/multipart"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/encoding"
	"github.com/go-seidon/provider/http"
	"github.com/go-seidon/provider/serialization"
	"github.com/go-seidon/provider/status"
	"github.com/go-seidon/provider/typeconv"
)

type hippoStorage struct {
	auth       *StorageAuth
	config     *StorageConfig
	fileWriter multipart.Writer
	serializer serialization.Serializer
	encoder    encoding.Encoder
	httpClient http.Client
	clock      datetime.Clock
}

func (s *hippoStorage) UploadObject(ctx context.Context, p storage.UploadObjectParam) (*storage.UploadObjectResult, error) {
	basicAuth, err := s.encoder.Encode([]byte(s.auth.ClientId + ":" + s.auth.ClientSecret))
	if err != nil {
		return nil, err
	}

	fileName := typeconv.StringVal(p.Name)
	if p.Extension != nil {
		fileName = fmt.Sprintf("%s.%s", fileName, typeconv.StringVal(p.Extension))
	}

	body := bytes.NewBuffer([]byte{})
	writer, err := s.fileWriter(multipart.WriterParam{
		Writer:    body,
		Reader:    p.Data,
		FieldName: "file",
		FileName:  fileName,
	})
	if err != nil {
		return nil, err
	}

	uploadObject, err := s.httpClient.Post(ctx, http.RequestParam{
		Url:  fmt.Sprintf("%s/%s", s.config.Host, "v1/file"),
		Body: body,
		Header: map[string][]string{
			"Content-Type":  {writer.FormDataContentType()},
			"Authorization": {fmt.Sprintf("%s %s", "Basic", basicAuth)},
		},
	})
	if err != nil {
		return nil, err
	}

	uploadData, err := io.ReadAll(uploadObject.Body)
	if err != nil {
		return nil, err
	}

	uploadBody := &UploadObjectResponseBody{}
	err = s.serializer.Unmarshal(uploadData, uploadBody)
	if err != nil {
		return nil, err
	}

	if uploadObject.StatusCode != net_http.StatusOK {
		if uploadBody.Code == status.ACTION_FORBIDDEN {
			return nil, storage.ErrUnauthenticated
		}
		return nil, fmt.Errorf(uploadBody.Message)
	}

	res := &storage.UploadObjectResult{
		ObjectId:   uploadBody.Data.Id,
		UploadedAt: time.UnixMilli(uploadBody.Data.UploadedAt).UTC(),
	}
	return res, nil
}

func (s *hippoStorage) RetrieveObject(ctx context.Context, p storage.RetrieveObjectParam) (*storage.RetrieveObjectResult, error) {
	basicAuth, err := s.encoder.Encode([]byte(s.auth.ClientId + ":" + s.auth.ClientSecret))
	if err != nil {
		return nil, err
	}

	retrieveObject, err := s.httpClient.Get(ctx, http.RequestParam{
		Url: fmt.Sprintf("%s/%s/%s", s.config.Host, "v1/file", p.ObjectId),
		Header: map[string][]string{
			"Authorization": {fmt.Sprintf("%s %s", "Basic", basicAuth)},
		},
	})
	if err != nil {
		return nil, err
	}

	if retrieveObject.StatusCode != net_http.StatusOK {
		retrieveData, err := io.ReadAll(retrieveObject.Body)
		if err != nil {
			return nil, err
		}

		retrieveBody := &ResponseBodyInfo{}
		err = s.serializer.Unmarshal(retrieveData, retrieveBody)
		if err != nil {
			return nil, err
		}

		err = fmt.Errorf(retrieveBody.Message)
		switch retrieveBody.Code {
		case status.ACTION_FORBIDDEN:
			err = storage.ErrUnauthenticated
		case status.RESOURCE_NOTFOUND:
			err = storage.ErrNotFound
		}
		return nil, err
	}

	res := &storage.RetrieveObjectResult{
		Data:        retrieveObject.Body,
		RetrievedAt: s.clock.Now().UTC(),
	}
	return res, nil
}

func NewStorage(opts ...StorageOption) *hippoStorage {
	p := StorageParam{}
	for _, opt := range opts {
		opt(&p)
	}

	return &hippoStorage{
		auth:       p.Auth,
		config:     p.Config,
		fileWriter: p.FileWriter,
		serializer: p.Serializer,
		encoder:    p.Encoder,
		httpClient: p.HttpClient,
		clock:      p.Clock,
	}
}
