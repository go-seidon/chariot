package hippo_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	mime_multipart "mime/multipart"
	"testing"
	"time"

	"github.com/go-seidon/chariot/internal/storage"
	"github.com/go-seidon/chariot/internal/storage/hippo"
	"github.com/go-seidon/chariot/internal/storage/multipart"
	mock_encoding "github.com/go-seidon/provider/encoding/mock"
	"github.com/go-seidon/provider/http"
	mock_http "github.com/go-seidon/provider/http/mock"
	mock_io "github.com/go-seidon/provider/io/mock"
	"github.com/go-seidon/provider/serialization/json"
	mock_serialization "github.com/go-seidon/provider/serialization/mock"
	"github.com/go-seidon/provider/typeconv"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHippo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hippo Package")
}

var _ = Describe("Hippo Storage", func() {

	Context("UploadObject function", Label("unit"), func() {
		var (
			ctx        context.Context
			ctrl       *gomock.Controller
			currentTs  time.Time
			s          storage.Storage
			file       *mock_io.MockReader
			encoder    *mock_encoding.MockEncoder
			basicAuth  string
			auth       *hippo.StorageAuth
			config     *hippo.StorageConfig
			writer     multipart.Writer
			httpClient *mock_http.MockClient
			serializer *mock_serialization.MockSerializer
			p          storage.UploadObjectParam
			uploadRes  *http.RequestResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			t := GinkgoT()
			currentTs = time.Now().UTC()
			ctrl = gomock.NewController(t)
			encoder = mock_encoding.NewMockEncoder(ctrl)
			auth = &hippo.StorageAuth{
				ClientId:     "client-id",
				ClientSecret: "client-secret",
			}
			config = &hippo.StorageConfig{
				Host: "host",
			}
			writer = func(p multipart.WriterParam) (*mime_multipart.Writer, error) {
				return &mime_multipart.Writer{}, nil
			}
			httpClient = mock_http.NewMockClient(ctrl)
			serializer = mock_serialization.NewMockSerializer(ctrl)
			s = hippo.NewStorage(
				hippo.WithAuth(auth),
				hippo.WithEncoder(encoder),
				hippo.WithConfig(config),
				hippo.WithWriter(writer),
				hippo.WithHttpClient(httpClient),
				hippo.WithSerializer(serializer),
			)
			file = mock_io.NewMockReader(ctrl)
			basicAuth = "auth-token"
			p = storage.UploadObjectParam{
				Data:      file,
				Id:        typeconv.String("object-id"),
				Extension: typeconv.String("jpg"),
			}
			response := &hippo.UploadObjectResponseBody{
				Code:    1000,
				Message: "success upload file",
				Data: hippo.UploadObjectResponseData{
					Id:         "object-id",
					UploadedAt: currentTs.UnixMilli(),
				},
			}
			js := json.NewSerializer()
			body, _ := js.Marshal(response)
			buffer := bytes.NewBuffer(body)
			rc := io.NopCloser(buffer)
			uploadRes = &http.RequestResult{
				StatusCode: 200,
				Body:       rc,
			}
		})

		When("failed encode auth", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Encode(gomock.Eq([]byte(fmt.Sprintf("%s:%s", auth.ClientId, auth.ClientSecret)))).
					Return("", fmt.Errorf("disk error")).
					Times(1)

				res, err := s.UploadObject(ctx, p)

				Expect(err).To(Equal(fmt.Errorf("disk error")))
				Expect(res).To(BeNil())
			})
		})

		When("failed create file writer", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Encode(gomock.Eq([]byte(fmt.Sprintf("%s:%s", auth.ClientId, auth.ClientSecret)))).
					Return(basicAuth, nil).
					Times(1)

				writer = func(p multipart.WriterParam) (*mime_multipart.Writer, error) {
					return nil, fmt.Errorf("disk error")
				}
				s := hippo.NewStorage(
					hippo.WithAuth(auth),
					hippo.WithEncoder(encoder),
					hippo.WithConfig(config),
					hippo.WithWriter(writer),
				)

				res, err := s.UploadObject(ctx, p)

				Expect(err).To(Equal(fmt.Errorf("disk error")))
				Expect(res).To(BeNil())
			})
		})

		When("failed execute http request", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Encode(gomock.Eq([]byte(fmt.Sprintf("%s:%s", auth.ClientId, auth.ClientSecret)))).
					Return(basicAuth, nil).
					Times(1)

				httpClient.
					EXPECT().
					Post(gomock.Eq(ctx), gomock.Any()).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := s.UploadObject(ctx, p)

				Expect(err).To(Equal(fmt.Errorf("network error")))
				Expect(res).To(BeNil())
			})
		})

		When("failed read upload response", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Encode(gomock.Eq([]byte(fmt.Sprintf("%s:%s", auth.ClientId, auth.ClientSecret)))).
					Return(basicAuth, nil).
					Times(1)

				body := mock_io.NewMockReadCloser(ctrl)
				body.
					EXPECT().
					Read(gomock.Any()).
					Return(0, fmt.Errorf("read error")).
					Times(1)

				uploadRes := &http.RequestResult{
					StatusCode: 200,
					Body:       body,
				}
				httpClient.
					EXPECT().
					Post(gomock.Eq(ctx), gomock.Any()).
					Return(uploadRes, nil).
					Times(1)

				res, err := s.UploadObject(ctx, p)

				Expect(err).To(Equal(fmt.Errorf("read error")))
				Expect(res).To(BeNil())
			})
		})

		When("failed unmarshall response", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Encode(gomock.Eq([]byte(fmt.Sprintf("%s:%s", auth.ClientId, auth.ClientSecret)))).
					Return(basicAuth, nil).
					Times(1)

				httpClient.
					EXPECT().
					Post(gomock.Eq(ctx), gomock.Any()).
					Return(uploadRes, nil).
					Times(1)

				serializer.
					EXPECT().
					Unmarshal(gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("disk error")).
					Times(1)

				res, err := s.UploadObject(ctx, p)

				Expect(err).To(Equal(fmt.Errorf("disk error")))
				Expect(res).To(BeNil())
			})
		})

		When("not allowed to upload", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Encode(gomock.Eq([]byte(fmt.Sprintf("%s:%s", auth.ClientId, auth.ClientSecret)))).
					Return(basicAuth, nil).
					Times(1)

				response := &hippo.UploadObjectResponseBody{
					Code:    1003,
					Message: "unauthenticated access",
				}
				js := json.NewSerializer()
				body, _ := js.Marshal(response)
				buffer := bytes.NewBuffer(body)
				rc := io.NopCloser(buffer)
				uploadRes = &http.RequestResult{
					StatusCode: 403,
					Body:       rc,
				}
				httpClient.
					EXPECT().
					Post(gomock.Eq(ctx), gomock.Any()).
					Return(uploadRes, nil).
					Times(1)

				serializer := json.NewSerializer()
				s := hippo.NewStorage(
					hippo.WithAuth(auth),
					hippo.WithEncoder(encoder),
					hippo.WithConfig(config),
					hippo.WithWriter(writer),
					hippo.WithHttpClient(httpClient),
					hippo.WithSerializer(serializer),
				)

				res, err := s.UploadObject(ctx, p)

				Expect(err).To(Equal(storage.ErrUnauthenticated))
				Expect(res).To(BeNil())
			})
		})

		When("failed upload object", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Encode(gomock.Eq([]byte(fmt.Sprintf("%s:%s", auth.ClientId, auth.ClientSecret)))).
					Return(basicAuth, nil).
					Times(1)

				response := &hippo.UploadObjectResponseBody{
					Code:    1001,
					Message: "database error",
				}
				js := json.NewSerializer()
				body, _ := js.Marshal(response)
				buffer := bytes.NewBuffer(body)
				rc := io.NopCloser(buffer)
				uploadRes = &http.RequestResult{
					StatusCode: 500,
					Body:       rc,
				}
				httpClient.
					EXPECT().
					Post(gomock.Eq(ctx), gomock.Any()).
					Return(uploadRes, nil).
					Times(1)

				serializer := json.NewSerializer()
				s := hippo.NewStorage(
					hippo.WithAuth(auth),
					hippo.WithEncoder(encoder),
					hippo.WithConfig(config),
					hippo.WithWriter(writer),
					hippo.WithHttpClient(httpClient),
					hippo.WithSerializer(serializer),
				)

				res, err := s.UploadObject(ctx, p)

				Expect(err).To(Equal(fmt.Errorf("database error")))
				Expect(res).To(BeNil())
			})
		})

		When("success upload object", func() {
			It("should return error", func() {
				encoder.
					EXPECT().
					Encode(gomock.Eq([]byte(fmt.Sprintf("%s:%s", auth.ClientId, auth.ClientSecret)))).
					Return(basicAuth, nil).
					Times(1)

				httpClient.
					EXPECT().
					Post(gomock.Eq(ctx), gomock.Any()).
					Return(uploadRes, nil).
					Times(1)

				serializer := json.NewSerializer()
				s := hippo.NewStorage(
					hippo.WithAuth(auth),
					hippo.WithEncoder(encoder),
					hippo.WithConfig(config),
					hippo.WithWriter(writer),
					hippo.WithHttpClient(httpClient),
					hippo.WithSerializer(serializer),
				)

				res, err := s.UploadObject(ctx, p)

				Expect(err).To(BeNil())
				Expect(res.ObjectId).To(Equal("object-id"))
				Expect(res.UploadedAt).To(Equal(time.UnixMilli(currentTs.UnixMilli()).UTC()))
			})
		})
	})
})
