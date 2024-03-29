package service_test

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-seidon/chariot/api/queue"
	"github.com/go-seidon/chariot/internal/repository"
	mock_repository "github.com/go-seidon/chariot/internal/repository/mock"
	"github.com/go-seidon/chariot/internal/service"
	mock_service "github.com/go-seidon/chariot/internal/service/mock"
	"github.com/go-seidon/chariot/internal/storage"
	mock_storage "github.com/go-seidon/chariot/internal/storage/mock"
	"github.com/go-seidon/chariot/internal/storage/router"
	mock_datetime "github.com/go-seidon/provider/datetime/mock"
	mock_identifier "github.com/go-seidon/provider/identity/mock"
	mock_io "github.com/go-seidon/provider/io/mock"
	"github.com/go-seidon/provider/queueing"
	mock_queueing "github.com/go-seidon/provider/queueing/mock"
	mock_serialization "github.com/go-seidon/provider/serialization/mock"
	mock_slug "github.com/go-seidon/provider/slug/mock"
	"github.com/go-seidon/provider/system"
	"github.com/go-seidon/provider/typeconv"
	mock_validation "github.com/go-seidon/provider/validation/mock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("File Package", func() {
	Context("UploadFile function", Label("unit"), func() {
		var (
			ctx             context.Context
			currentTs       time.Time
			fileClient      service.File
			validator       *mock_validation.MockValidator
			identifier      *mock_identifier.MockIdentifier
			clock           *mock_datetime.MockClock
			slugger         *mock_slug.MockSlugger
			barrelRepo      *mock_repository.MockBarrel
			fileRepo        *mock_repository.MockFile
			sessionClient   *mock_service.MockSession
			storageRouter   *mock_storage.MockRouter
			storagePrimary  *mock_storage.MockStorage
			fileData        *mock_io.MockReader
			p               service.UploadFileParam
			r               *service.UploadFileResult
			createSessParam service.CreateSessionParam
			createSessRes   *service.CreateSessionResult
			searchParam     repository.SearchBarrelParam
			searchRes       *repository.SearchBarrelResult
			createStgParam  router.CreateStorageParam
			uploadParam     storage.UploadObjectParam
			uploadRes       *storage.UploadObjectResult
			createFileParam repository.CreateFileParam
			createFileRes   *repository.CreateFileResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			validator = mock_validation.NewMockValidator(ctrl)
			identifier = mock_identifier.NewMockIdentifier(ctrl)
			clock = mock_datetime.NewMockClock(ctrl)
			slugger = mock_slug.NewMockSlugger(ctrl)
			barrelRepo = mock_repository.NewMockBarrel(ctrl)
			fileRepo = mock_repository.NewMockFile(ctrl)
			storageRouter = mock_storage.NewMockRouter(ctrl)
			storagePrimary = mock_storage.NewMockStorage(ctrl)
			fileData = mock_io.NewMockReader(ctrl)
			sessionClient = mock_service.NewMockSession(ctrl)
			fileClient = service.NewFile(service.FileParam{
				Config: &service.FileConfig{
					AppHost: "http://localhost",
				},
				Validator:     validator,
				Identifier:    identifier,
				Clock:         clock,
				Slugger:       slugger,
				BarrelRepo:    barrelRepo,
				FileRepo:      fileRepo,
				Router:        storageRouter,
				SessionClient: sessionClient,
			})
			p = service.UploadFileParam{
				Data: fileData,
				Info: service.UploadFileInfo{
					Name:      "Dolphin 22",
					Mimetype:  "image/jpeg",
					Extension: "jpg",
					Size:      23343,
					Meta: map[string]string{
						"feature": "profile",
						"user_id": "8c7ffa05-70c7-437e-8166-0f6a651a9575",
					},
				},
				Setting: service.UploadFileSetting{
					Visibility: "public",
					Barrels:    []string{"hippo1", "s3backup"},
				},
			}
			createSessParam = service.CreateSessionParam{
				Duration: 1800,
				Features: []string{"retrieve_file"},
			}
			createSessRes = &service.CreateSessionResult{
				Success:   system.Success{},
				CreatedAt: currentTs.UTC(),
				ExpiresAt: currentTs.Add(1800 * time.Second).UTC(),
				Token:     "secret-token",
			}
			searchParam = repository.SearchBarrelParam{
				Codes:    []string{"hippo1", "s3backup"},
				Statuses: []string{"active"},
			}
			searchRes = &repository.SearchBarrelResult{
				Summary: repository.SearchBarrelSummary{
					TotalItems: 2,
				},
				Items: []repository.SearchBarrelItem{
					{
						Id:   "s1",
						Code: "s3backup",
					},
					{
						Id:   "h1",
						Code: "hippo1",
					},
				},
			}
			createStgParam = router.CreateStorageParam{
				BarrelCode: "hippo1",
			}
			uploadParam = storage.UploadObjectParam{
				Data:      p.Data,
				Id:        typeconv.String("mock-id"),
				Name:      typeconv.String(strings.ToLower(p.Info.Name)),
				Mimetype:  typeconv.String(strings.ToLower(p.Info.Mimetype)),
				Extension: typeconv.String(strings.ToLower(p.Info.Extension)),
			}
			uploadRes = &storage.UploadObjectResult{
				ObjectId:   "object-id",
				UploadedAt: currentTs,
			}
			createFileParam = repository.CreateFileParam{
				Id:         "mock-id",
				Slug:       "dolphin-22.jpg",
				Name:       strings.ToLower(p.Info.Name),
				Mimetype:   strings.ToLower(p.Info.Mimetype),
				Extension:  strings.ToLower(p.Info.Extension),
				Size:       p.Info.Size,
				Visibility: p.Setting.Visibility,
				Status:     "available",
				Meta:       p.Info.Meta,
				CreatedAt:  currentTs,
				UploadedAt: currentTs,
				Locations: []repository.CreateFileLocation{
					{
						Id:         "mock-id",
						BarrelId:   "h1",
						ExternalId: typeconv.String("object-id"),
						Priority:   1,
						CreatedAt:  currentTs,
						Status:     "available",
						UploadedAt: &currentTs,
					},
					{
						Id:         "mock-id",
						BarrelId:   "s1",
						ExternalId: nil,
						Priority:   2,
						CreatedAt:  currentTs,
						Status:     "pending",
						UploadedAt: nil,
					},
				},
			}
			createFileRes = &repository.CreateFileResult{
				Id:         createFileParam.Id,
				Slug:       createFileParam.Slug,
				Name:       createFileParam.Name,
				Mimetype:   createFileParam.Mimetype,
				Extension:  createFileParam.Extension,
				Size:       createFileParam.Size,
				Visibility: createFileParam.Visibility,
				Status:     createFileParam.Status,
				Meta:       createFileParam.Meta,
				CreatedAt:  createFileParam.CreatedAt,
				UploadedAt: createFileParam.UploadedAt,
			}
			r = &service.UploadFileResult{
				Success: system.Success{
					Code:    1000,
					Message: "success upload file",
				},
				Id:         createFileRes.Id,
				Slug:       createFileRes.Slug,
				Name:       createFileRes.Name,
				Mimetype:   createFileRes.Mimetype,
				Extension:  createFileRes.Extension,
				Size:       createFileRes.Size,
				Visibility: createFileRes.Visibility,
				Status:     createFileRes.Status,
				FileUrl:    fmt.Sprintf("%s/file/%s", "http://localhost", createFileRes.Slug),
				AccessUrl:  fmt.Sprintf("%s/file/%s", "http://localhost", createFileRes.Slug),
				UploadedAt: createFileRes.UploadedAt,
				Meta:       createFileRes.Meta,
			}
		})

		When("file data is not specified", func() {
			It("should return error", func() {
				p.Data = nil

				res, err := fileClient.UploadFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("file is not specified"))
			})
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(fmt.Errorf("invalid data")).
					Times(1)

				res, err := fileClient.UploadFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("invalid data"))
			})
		})

		When("failed create session", func() {
			It("should return error", func() {
				p := service.UploadFileParam{
					Data: fileData,
					Info: service.UploadFileInfo{
						Name:      "Dolphin 22",
						Mimetype:  "image/jpeg",
						Extension: "jpg",
						Size:      23343,
						Meta: map[string]string{
							"feature": "profile",
							"user_id": "8c7ffa05-70c7-437e-8166-0f6a651a9575",
						},
					},
					Setting: service.UploadFileSetting{
						Visibility: "protected",
						Barrels:    []string{"hippo1", "s3backup"},
					},
				}
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				sessionClient.
					EXPECT().
					CreateSession(gomock.Eq(ctx), gomock.Eq(createSessParam)).
					Return(nil, &system.Error{
						Code:    1001,
						Message: "disk error",
					}).
					Times(1)

				res, err := fileClient.UploadFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("disk error"))
			})
		})

		When("failed search barrels", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				barrelRepo.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.UploadFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("there is invalid barrel", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				searchRes := &repository.SearchBarrelResult{
					Summary: repository.SearchBarrelSummary{},
					Items:   []repository.SearchBarrelItem{},
				}
				barrelRepo.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				res, err := fileClient.UploadFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("there is invalid barrel"))
			})
		})

		When("failed generate id", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				barrelRepo.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("mock-id", nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("", fmt.Errorf("disk error")).
					Times(1)

				res, err := fileClient.UploadFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("disk error"))
			})
		})

		When("storage is not supported", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				barrelRepo.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("mock-id", nil).
					Times(2)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(nil, router.ErrUnsupported).
					Times(1)

				res, err := fileClient.UploadFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("unsupported storage"))
			})
		})

		When("failed upload file", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				barrelRepo.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("mock-id", nil).
					Times(2)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(storagePrimary, nil).
					Times(1)

				storagePrimary.
					EXPECT().
					UploadObject(gomock.Eq(ctx), gomock.Eq(uploadParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.UploadFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("failed create file", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				barrelRepo.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("mock-id", nil).
					Times(2)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(storagePrimary, nil).
					Times(1)

				storagePrimary.
					EXPECT().
					UploadObject(gomock.Eq(ctx), gomock.Eq(uploadParam)).
					Return(uploadRes, nil).
					Times(1)

				slugger.
					EXPECT().
					GenerateSlug(gomock.Eq(p.Info.Name)).
					Return("dolphin-22").
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					CreateFile(gomock.Eq(ctx), gomock.Eq(createFileParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.UploadFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("success upload file", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				barrelRepo.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("mock-id", nil).
					Times(2)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(storagePrimary, nil).
					Times(1)

				storagePrimary.
					EXPECT().
					UploadObject(gomock.Eq(ctx), gomock.Eq(uploadParam)).
					Return(uploadRes, nil).
					Times(1)

				slugger.
					EXPECT().
					GenerateSlug(gomock.Eq(p.Info.Name)).
					Return("dolphin-22").
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					CreateFile(gomock.Eq(ctx), gomock.Eq(createFileParam)).
					Return(createFileRes, nil).
					Times(1)

				res, err := fileClient.UploadFile(ctx, p)

				Expect(err).To(BeNil())
				Expect(res).To(Equal(r))
			})
		})

		When("success upload to one barrel", func() {
			It("should return result", func() {
				p := service.UploadFileParam{
					Data: fileData,
					Info: service.UploadFileInfo{
						Name:      "Dolphin 22",
						Mimetype:  "image/jpeg",
						Extension: "",
						Size:      23343,
						Meta: map[string]string{
							"feature": "profile",
							"user_id": "8c7ffa05-70c7-437e-8166-0f6a651a9575",
						},
					},
					Setting: service.UploadFileSetting{
						Visibility: "public",
						Barrels:    []string{"hippo1"},
					},
				}
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				searchParam := repository.SearchBarrelParam{
					Codes:    []string{"hippo1"},
					Statuses: []string{"active"},
				}
				searchRes = &repository.SearchBarrelResult{
					Summary: repository.SearchBarrelSummary{
						TotalItems: 1,
					},
					Items: []repository.SearchBarrelItem{
						{
							Id:   "h1",
							Code: "hippo1",
						},
					},
				}
				barrelRepo.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("mock-id", nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(storagePrimary, nil).
					Times(1)

				uploadParam := storage.UploadObjectParam{
					Data:      p.Data,
					Id:        typeconv.String("mock-id"),
					Name:      typeconv.String(strings.ToLower(p.Info.Name)),
					Mimetype:  typeconv.String(strings.ToLower(p.Info.Mimetype)),
					Extension: typeconv.String(strings.ToLower(p.Info.Extension)),
				}
				uploadRes := &storage.UploadObjectResult{
					ObjectId:   "object-id",
					UploadedAt: currentTs,
				}
				storagePrimary.
					EXPECT().
					UploadObject(gomock.Eq(ctx), gomock.Eq(uploadParam)).
					Return(uploadRes, nil).
					Times(1)

				slugger.
					EXPECT().
					GenerateSlug(gomock.Eq(p.Info.Name)).
					Return("dolphin-22").
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				createFileParam = repository.CreateFileParam{
					Id:         "mock-id",
					Slug:       "dolphin-22",
					Name:       strings.ToLower(p.Info.Name),
					Mimetype:   strings.ToLower(p.Info.Mimetype),
					Extension:  strings.ToLower(p.Info.Extension),
					Size:       p.Info.Size,
					Visibility: p.Setting.Visibility,
					Status:     "available",
					Meta:       p.Info.Meta,
					CreatedAt:  currentTs,
					UploadedAt: currentTs,
					Locations: []repository.CreateFileLocation{
						{
							Id:         "mock-id",
							BarrelId:   "h1",
							ExternalId: typeconv.String("object-id"),
							Priority:   1,
							CreatedAt:  currentTs,
							Status:     "available",
							UploadedAt: &currentTs,
						},
					},
				}
				createFileRes = &repository.CreateFileResult{
					Id:         createFileParam.Id,
					Slug:       createFileParam.Slug,
					Name:       createFileParam.Name,
					Mimetype:   createFileParam.Mimetype,
					Extension:  createFileParam.Extension,
					Size:       createFileParam.Size,
					Visibility: createFileParam.Visibility,
					Status:     createFileParam.Status,
					Meta:       createFileParam.Meta,
					CreatedAt:  createFileParam.CreatedAt,
					UploadedAt: createFileParam.UploadedAt,
				}
				fileRepo.
					EXPECT().
					CreateFile(gomock.Eq(ctx), gomock.Eq(createFileParam)).
					Return(createFileRes, nil).
					Times(1)

				res, err := fileClient.UploadFile(ctx, p)

				r := &service.UploadFileResult{
					Success: system.Success{
						Code:    1000,
						Message: "success upload file",
					},
					Id:         createFileRes.Id,
					Slug:       createFileRes.Slug,
					Name:       createFileRes.Name,
					Mimetype:   createFileRes.Mimetype,
					Extension:  createFileRes.Extension,
					Size:       createFileRes.Size,
					Visibility: createFileRes.Visibility,
					Status:     createFileRes.Status,
					FileUrl:    fmt.Sprintf("%s/file/%s", "http://localhost", createFileRes.Slug),
					AccessUrl:  fmt.Sprintf("%s/file/%s", "http://localhost", createFileRes.Slug),
					UploadedAt: createFileRes.UploadedAt,
					Meta:       createFileRes.Meta,
				}
				Expect(err).To(BeNil())
				Expect(res).To(Equal(r))
			})
		})

		When("success upload protected file", func() {
			It("should return result", func() {
				p := service.UploadFileParam{
					Data: fileData,
					Info: service.UploadFileInfo{
						Name:      "Dolphin 22",
						Mimetype:  "image/jpeg",
						Extension: "jpg",
						Size:      23343,
						Meta: map[string]string{
							"feature": "profile",
							"user_id": "8c7ffa05-70c7-437e-8166-0f6a651a9575",
						},
					},
					Setting: service.UploadFileSetting{
						Visibility: "protected",
						Barrels:    []string{"hippo1", "s3backup"},
					},
				}
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				sessionClient.
					EXPECT().
					CreateSession(gomock.Eq(ctx), gomock.Eq(createSessParam)).
					Return(createSessRes, nil).
					Times(1)

				barrelRepo.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("mock-id", nil).
					Times(2)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(storagePrimary, nil).
					Times(1)

				storagePrimary.
					EXPECT().
					UploadObject(gomock.Eq(ctx), gomock.Eq(uploadParam)).
					Return(uploadRes, nil).
					Times(1)

				slugger.
					EXPECT().
					GenerateSlug(gomock.Eq(p.Info.Name)).
					Return("dolphin-22").
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				createFileParam := repository.CreateFileParam{
					Id:         "mock-id",
					Slug:       "dolphin-22.jpg",
					Name:       strings.ToLower(p.Info.Name),
					Mimetype:   strings.ToLower(p.Info.Mimetype),
					Extension:  strings.ToLower(p.Info.Extension),
					Size:       p.Info.Size,
					Visibility: p.Setting.Visibility,
					Status:     "available",
					Meta:       p.Info.Meta,
					CreatedAt:  currentTs,
					UploadedAt: currentTs,
					Locations: []repository.CreateFileLocation{
						{
							Id:         "mock-id",
							BarrelId:   "h1",
							ExternalId: typeconv.String("object-id"),
							Priority:   1,
							CreatedAt:  currentTs,
							Status:     "available",
							UploadedAt: &currentTs,
						},
						{
							Id:         "mock-id",
							BarrelId:   "s1",
							ExternalId: nil,
							Priority:   2,
							CreatedAt:  currentTs,
							Status:     "pending",
							UploadedAt: nil,
						},
					},
				}
				createFileRes := &repository.CreateFileResult{
					Id:         createFileParam.Id,
					Slug:       createFileParam.Slug,
					Name:       createFileParam.Name,
					Mimetype:   createFileParam.Mimetype,
					Extension:  createFileParam.Extension,
					Size:       createFileParam.Size,
					Visibility: createFileParam.Visibility,
					Status:     createFileParam.Status,
					Meta:       createFileParam.Meta,
					CreatedAt:  createFileParam.CreatedAt,
					UploadedAt: createFileParam.UploadedAt,
				}
				fileRepo.
					EXPECT().
					CreateFile(gomock.Eq(ctx), gomock.Eq(createFileParam)).
					Return(createFileRes, nil).
					Times(1)

				res, err := fileClient.UploadFile(ctx, p)

				r := &service.UploadFileResult{
					Success: system.Success{
						Code:    1000,
						Message: "success upload file",
					},
					Id:         createFileRes.Id,
					Slug:       createFileRes.Slug,
					Name:       createFileRes.Name,
					Mimetype:   createFileRes.Mimetype,
					Extension:  createFileRes.Extension,
					Size:       createFileRes.Size,
					Visibility: createFileRes.Visibility,
					Status:     createFileRes.Status,
					FileUrl:    fmt.Sprintf("%s/file/%s", "http://localhost", createFileRes.Slug),
					AccessUrl:  fmt.Sprintf("%s/file/%s?token=%s", "http://localhost", createFileRes.Slug, createSessRes.Token),
					UploadedAt: createFileRes.UploadedAt,
					Meta:       createFileRes.Meta,
				}
				Expect(err).To(BeNil())
				Expect(res).To(Equal(r))
			})
		})
	})

	Context("RetrieveFileBySlug function", Label("unit"), func() {
		var (
			ctx            context.Context
			currentTs      time.Time
			fileClient     service.File
			validator      *mock_validation.MockValidator
			identifier     *mock_identifier.MockIdentifier
			clock          *mock_datetime.MockClock
			slugger        *mock_slug.MockSlugger
			barrelRepo     *mock_repository.MockBarrel
			fileRepo       *mock_repository.MockFile
			storageRouter  *mock_storage.MockRouter
			storagePrimary *mock_storage.MockStorage
			sessionClient  *mock_service.MockSession
			fileData       *mock_io.MockReadCloser
			p              service.RetrieveFileBySlugParam
			r              *service.RetrieveFileBySlugResult
			createStgParam router.CreateStorageParam
			retrieveParam  storage.RetrieveObjectParam
			retrieveRes    *storage.RetrieveObjectResult
			findFileParam  repository.FindFileParam
			findFileRes    *repository.FindFileResult
			verifyParam    service.VerifySessionParam
			verifyRes      *service.VerifySessionResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			validator = mock_validation.NewMockValidator(ctrl)
			identifier = mock_identifier.NewMockIdentifier(ctrl)
			clock = mock_datetime.NewMockClock(ctrl)
			slugger = mock_slug.NewMockSlugger(ctrl)
			barrelRepo = mock_repository.NewMockBarrel(ctrl)
			fileRepo = mock_repository.NewMockFile(ctrl)
			storageRouter = mock_storage.NewMockRouter(ctrl)
			storagePrimary = mock_storage.NewMockStorage(ctrl)
			sessionClient = mock_service.NewMockSession(ctrl)
			fileData = mock_io.NewMockReadCloser(ctrl)
			fileClient = service.NewFile(service.FileParam{
				Validator:     validator,
				Identifier:    identifier,
				Clock:         clock,
				Slugger:       slugger,
				BarrelRepo:    barrelRepo,
				FileRepo:      fileRepo,
				Router:        storageRouter,
				SessionClient: sessionClient,
			})

			createStgParam = router.CreateStorageParam{
				BarrelCode: "b1",
			}
			retrieveParam = storage.RetrieveObjectParam{
				ObjectId: "e1",
			}
			retrieveRes = &storage.RetrieveObjectResult{
				Data:        fileData,
				RetrievedAt: currentTs,
			}
			p = service.RetrieveFileBySlugParam{
				Slug:  "dolphin-22.jpg",
				Token: "session-token",
			}
			findFileParam = repository.FindFileParam{
				Slug: p.Slug,
			}
			findFileRes = &repository.FindFileResult{
				Id:         "id",
				Status:     "available",
				Visibility: "public",
				Locations: []repository.FindFileLocation{
					{
						Barrel: repository.FindFileBarrel{
							Id:       "b1",
							Code:     "b1",
							Provider: "goseidon_hippo",
							Status:   "active",
						},
						ExternalId: typeconv.String("e1"),
						Priority:   1,
						Status:     "available",
						CreatedAt:  currentTs,
						UpdatedAt:  typeconv.Time(currentTs),
						UploadedAt: typeconv.Time(currentTs),
					},
					{
						Barrel: repository.FindFileBarrel{
							Id:       "b2",
							Code:     "b2",
							Provider: "goseidon_hippo",
							Status:   "active",
						},
						ExternalId: typeconv.String("e2"),
						Priority:   2,
						Status:     "uploading",
						CreatedAt:  currentTs,
						UpdatedAt:  typeconv.Time(currentTs),
						UploadedAt: nil,
					},
				},
			}
			r = &service.RetrieveFileBySlugResult{
				Success: system.Success{
					Code:    1000,
					Message: "success retrieve file",
				},
				Data:       fileData,
				Id:         findFileRes.Id,
				Visibility: findFileRes.Visibility,
				Status:     findFileRes.Status,
			}
			verifyParam = service.VerifySessionParam{
				Token:   p.Token,
				Feature: "retrieve_file",
			}
			verifyRes = &service.VerifySessionResult{}
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(fmt.Errorf("invalid data")).
					Times(1)

				res, err := fileClient.RetrieveFileBySlug(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("invalid data"))
			})
		})

		When("failed find file", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.RetrieveFileBySlug(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("file is not available", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(nil, repository.ErrNotFound).
					Times(1)

				res, err := fileClient.RetrieveFileBySlug(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1004)))
				Expect(err.Message).To(Equal("file is not available"))
			})
		})

		When("file status is not available", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				findFileRes = &repository.FindFileResult{
					Status: "deleting",
				}
				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				res, err := fileClient.RetrieveFileBySlug(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1004)))
				Expect(err.Message).To(Equal("file is not available"))
			})
		})

		When("failed verify session", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				findFileRes := &repository.FindFileResult{
					Id:         "id",
					Visibility: "protected",
					Status:     "available",
					Locations: []repository.FindFileLocation{
						{
							Barrel: repository.FindFileBarrel{
								Id:       "b1",
								Code:     "b1",
								Provider: "goseidon_hippo",
								Status:   "active",
							},
							ExternalId: typeconv.String("e1"),
							Priority:   1,
							Status:     "available",
							CreatedAt:  currentTs,
							UpdatedAt:  typeconv.Time(currentTs),
							UploadedAt: typeconv.Time(currentTs),
						},
						{
							Barrel: repository.FindFileBarrel{
								Id:       "b2",
								Code:     "b2",
								Provider: "goseidon_hippo",
								Status:   "active",
							},
							ExternalId: typeconv.String("e2"),
							Priority:   2,
							Status:     "uploading",
							CreatedAt:  currentTs,
							UpdatedAt:  typeconv.Time(currentTs),
							UploadedAt: nil,
						},
					},
				}
				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				sessionClient.
					EXPECT().
					VerifySession(gomock.Eq(ctx), gomock.Eq(verifyParam)).
					Return(nil, &system.Error{
						Code:    1001,
						Message: "i/o error",
					}).
					Times(1)

				res, err := fileClient.RetrieveFileBySlug(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1003)))
				Expect(err.Message).To(Equal("i/o error"))
			})
		})

		When("barrels are not active", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				findFileRes = &repository.FindFileResult{
					Status: "available",
					Locations: []repository.FindFileLocation{
						{
							Barrel: repository.FindFileBarrel{
								Status: "inactive",
							},
						},
					},
				}
				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				res, err := fileClient.RetrieveFileBySlug(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("barrels are not active"))
			})
		})

		When("data is invalid", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				findFileRes = &repository.FindFileResult{
					Status:    "available",
					Locations: []repository.FindFileLocation{},
				}
				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				res, err := fileClient.RetrieveFileBySlug(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("file data is invalid"))
			})
		})

		When("failed create storage", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(nil, fmt.Errorf("unsupported storage")).
					Times(1)

				res, err := fileClient.RetrieveFileBySlug(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("unsupported storage"))
			})
		})

		When("success retrieve object", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(storagePrimary, nil).
					Times(1)

				storagePrimary.
					EXPECT().
					RetrieveObject(gomock.Eq(ctx), gomock.Eq(retrieveParam)).
					Return(retrieveRes, nil).
					Times(1)

				res, err := fileClient.RetrieveFileBySlug(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("success retrieve object on last location", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				findFileRes := &repository.FindFileResult{
					Id:         "id",
					Visibility: "public",
					Status:     "available",
					Locations: []repository.FindFileLocation{
						{
							Barrel: repository.FindFileBarrel{
								Id:       "b1",
								Code:     "b1",
								Provider: "goseidon_hippo",
								Status:   "active",
							},
							ExternalId: typeconv.String("e1"),
							Priority:   1,
							Status:     "available",
							CreatedAt:  currentTs,
							UpdatedAt:  typeconv.Time(currentTs),
							UploadedAt: typeconv.Time(currentTs),
						},
						{
							Barrel: repository.FindFileBarrel{
								Id:       "b2",
								Code:     "b2",
								Provider: "goseidon_hippo",
								Status:   "active",
							},
							ExternalId: typeconv.String("e2"),
							Priority:   2,
							Status:     "available",
							CreatedAt:  currentTs,
							UpdatedAt:  typeconv.Time(currentTs),
							UploadedAt: typeconv.Time(currentTs),
						},
					},
				}
				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				createStgParam := router.CreateStorageParam{
					BarrelCode: "b1",
				}
				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(storagePrimary, nil).
					Times(1)

				retrieveParam := storage.RetrieveObjectParam{
					ObjectId: "e1",
				}
				storagePrimary.
					EXPECT().
					RetrieveObject(gomock.Eq(ctx), gomock.Eq(retrieveParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				createStgParam = router.CreateStorageParam{
					BarrelCode: "b2",
				}
				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(storagePrimary, nil).
					Times(1)

				retrieveParam = storage.RetrieveObjectParam{
					ObjectId: "e2",
				}
				storagePrimary.
					EXPECT().
					RetrieveObject(gomock.Eq(ctx), gomock.Eq(retrieveParam)).
					Return(retrieveRes, nil).
					Times(1)

				res, err := fileClient.RetrieveFileBySlug(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("failed retrieve object on last location", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				findFileRes := &repository.FindFileResult{
					Id:         "id",
					Visibility: "public",
					Status:     "available",
					Locations: []repository.FindFileLocation{
						{
							Barrel: repository.FindFileBarrel{
								Id:       "b1",
								Code:     "b1",
								Provider: "goseidon_hippo",
								Status:   "active",
							},
							ExternalId: typeconv.String("e1"),
							Priority:   1,
							Status:     "available",
							CreatedAt:  currentTs,
							UpdatedAt:  typeconv.Time(currentTs),
							UploadedAt: typeconv.Time(currentTs),
						},
						{
							Barrel: repository.FindFileBarrel{
								Id:       "b2",
								Code:     "b2",
								Provider: "goseidon_hippo",
								Status:   "active",
							},
							ExternalId: typeconv.String("e2"),
							Priority:   2,
							Status:     "available",
							CreatedAt:  currentTs,
							UpdatedAt:  typeconv.Time(currentTs),
							UploadedAt: typeconv.Time(currentTs),
						},
					},
				}
				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				createStgParam := router.CreateStorageParam{
					BarrelCode: "b1",
				}
				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(storagePrimary, nil).
					Times(1)

				retrieveParam := storage.RetrieveObjectParam{
					ObjectId: "e1",
				}
				storagePrimary.
					EXPECT().
					RetrieveObject(gomock.Eq(ctx), gomock.Eq(retrieveParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				createStgParam = router.CreateStorageParam{
					BarrelCode: "b2",
				}
				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(storagePrimary, nil).
					Times(1)

				retrieveParam = storage.RetrieveObjectParam{
					ObjectId: "e2",
				}
				storagePrimary.
					EXPECT().
					RetrieveObject(gomock.Eq(ctx), gomock.Eq(retrieveParam)).
					Return(nil, fmt.Errorf("i/o error")).
					Times(1)

				res, err := fileClient.RetrieveFileBySlug(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("failed retrieve file from barrel"))
			})
		})

		When("success retrieve protected object", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				findFileRes := &repository.FindFileResult{
					Id:         "id",
					Visibility: "protected",
					Status:     "available",
					Locations: []repository.FindFileLocation{
						{
							Barrel: repository.FindFileBarrel{
								Id:       "b1",
								Code:     "b1",
								Provider: "goseidon_hippo",
								Status:   "active",
							},
							ExternalId: typeconv.String("e1"),
							Priority:   1,
							Status:     "available",
							CreatedAt:  currentTs,
							UpdatedAt:  typeconv.Time(currentTs),
							UploadedAt: typeconv.Time(currentTs),
						},
						{
							Barrel: repository.FindFileBarrel{
								Id:       "b2",
								Code:     "b2",
								Provider: "goseidon_hippo",
								Status:   "active",
							},
							ExternalId: typeconv.String("e2"),
							Priority:   2,
							Status:     "uploading",
							CreatedAt:  currentTs,
							UpdatedAt:  typeconv.Time(currentTs),
							UploadedAt: nil,
						},
					},
				}
				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				sessionClient.
					EXPECT().
					VerifySession(gomock.Eq(ctx), gomock.Eq(verifyParam)).
					Return(verifyRes, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(storagePrimary, nil).
					Times(1)

				storagePrimary.
					EXPECT().
					RetrieveObject(gomock.Eq(ctx), gomock.Eq(retrieveParam)).
					Return(retrieveRes, nil).
					Times(1)

				res, err := fileClient.RetrieveFileBySlug(ctx, p)

				r := &service.RetrieveFileBySlugResult{
					Success: system.Success{
						Code:    1000,
						Message: "success retrieve file",
					},
					Data:       fileData,
					Id:         findFileRes.Id,
					Visibility: findFileRes.Visibility,
					Status:     findFileRes.Status,
				}
				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("replicas are unavailable", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				createStgParam := router.CreateStorageParam{
					BarrelCode: "b1",
				}
				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(storagePrimary, nil).
					Times(1)

				retrieveParam := storage.RetrieveObjectParam{
					ObjectId: "e1",
				}
				storagePrimary.
					EXPECT().
					RetrieveObject(gomock.Eq(ctx), gomock.Eq(retrieveParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.RetrieveFileBySlug(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("file replicas are not available"))
			})
		})
	})

	Context("GetFileById function", Label("unit"), func() {
		var (
			ctx           context.Context
			currentTs     time.Time
			fileClient    service.File
			validator     *mock_validation.MockValidator
			identifier    *mock_identifier.MockIdentifier
			clock         *mock_datetime.MockClock
			slugger       *mock_slug.MockSlugger
			barrelRepo    *mock_repository.MockBarrel
			fileRepo      *mock_repository.MockFile
			storageRouter *mock_storage.MockRouter
			p             service.GetFileByIdParam
			r             *service.GetFileByIdResult
			findFileParam repository.FindFileParam
			findFileRes   *repository.FindFileResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			validator = mock_validation.NewMockValidator(ctrl)
			identifier = mock_identifier.NewMockIdentifier(ctrl)
			clock = mock_datetime.NewMockClock(ctrl)
			slugger = mock_slug.NewMockSlugger(ctrl)
			barrelRepo = mock_repository.NewMockBarrel(ctrl)
			fileRepo = mock_repository.NewMockFile(ctrl)
			storageRouter = mock_storage.NewMockRouter(ctrl)
			fileClient = service.NewFile(service.FileParam{
				Validator:  validator,
				Identifier: identifier,
				Clock:      clock,
				Slugger:    slugger,
				BarrelRepo: barrelRepo,
				FileRepo:   fileRepo,
				Router:     storageRouter,
			})
			p = service.GetFileByIdParam{
				Id: "id",
			}
			findFileParam = repository.FindFileParam{
				Id: p.Id,
			}
			findFileRes = &repository.FindFileResult{
				Id:         "id",
				Slug:       "dolphin-22.jpg",
				Name:       "Dolphin 22",
				Mimetype:   "image/jpeg",
				Extension:  "jpg",
				Size:       23343,
				Visibility: "public",
				Status:     "available",
				UploadedAt: currentTs,
				CreatedAt:  currentTs,
				UpdatedAt:  &currentTs,
				DeletedAt:  nil,
				Meta: map[string]string{
					"feature": "profile",
					"user_id": "123",
				},
				Locations: []repository.FindFileLocation{
					{
						Barrel: repository.FindFileBarrel{
							Id:       "b1",
							Code:     "b1",
							Provider: "goseidon_hippo",
							Status:   "active",
						},
						ExternalId: typeconv.String("e1"),
						Priority:   1,
						Status:     "available",
						CreatedAt:  currentTs,
						UpdatedAt:  typeconv.Time(currentTs),
						UploadedAt: typeconv.Time(currentTs),
					},
					{
						Barrel: repository.FindFileBarrel{
							Id:       "b2",
							Code:     "b2",
							Provider: "goseidon_hippo",
							Status:   "active",
						},
						ExternalId: typeconv.String("e2"),
						Priority:   2,
						Status:     "uploading",
						CreatedAt:  currentTs,
						UpdatedAt:  typeconv.Time(currentTs),
						UploadedAt: nil,
					},
				},
			}

			locations := []service.GetFileByIdLocation{}
			for _, location := range findFileRes.Locations {
				locations = append(locations, service.GetFileByIdLocation{
					Barrel: service.GetFileByIdBarrel{
						Id:       location.Barrel.Id,
						Code:     location.Barrel.Code,
						Provider: location.Barrel.Provider,
						Status:   location.Barrel.Status,
					},
					ExternalId: location.ExternalId,
					Priority:   location.Priority,
					Status:     location.Status,
					CreatedAt:  location.CreatedAt,
					UpdatedAt:  location.UpdatedAt,
					UploadedAt: location.UploadedAt,
				})
			}
			r = &service.GetFileByIdResult{
				Success: system.Success{
					Code:    1000,
					Message: "success get file",
				},
				Id:         findFileRes.Id,
				Slug:       findFileRes.Slug,
				Name:       findFileRes.Name,
				Mimetype:   findFileRes.Mimetype,
				Extension:  findFileRes.Extension,
				Size:       findFileRes.Size,
				Visibility: findFileRes.Visibility,
				Status:     findFileRes.Status,
				UploadedAt: findFileRes.UploadedAt,
				CreatedAt:  findFileRes.CreatedAt,
				UpdatedAt:  findFileRes.UpdatedAt,
				DeletedAt:  findFileRes.DeletedAt,
				Meta:       findFileRes.Meta,
				Locations:  locations,
			}
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(fmt.Errorf("invalid data")).
					Times(1)

				res, err := fileClient.GetFileById(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("invalid data"))
			})
		})

		When("failed get file", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.GetFileById(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("file is not available", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(nil, repository.ErrNotFound).
					Times(1)

				res, err := fileClient.GetFileById(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1004)))
				Expect(err.Message).To(Equal("file is not available"))
			})
		})

		When("success get file", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				res, err := fileClient.GetFileById(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("SearchFile function", Label("unit"), func() {
		var (
			ctx         context.Context
			currentTs   time.Time
			fileClient  service.File
			p           service.SearchFileParam
			validator   *mock_validation.MockValidator
			identifier  *mock_identifier.MockIdentifier
			clock       *mock_datetime.MockClock
			fileRepo    *mock_repository.MockFile
			searchParam repository.SearchFileParam
			searchRes   *repository.SearchFileResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			validator = mock_validation.NewMockValidator(ctrl)
			identifier = mock_identifier.NewMockIdentifier(ctrl)
			clock = mock_datetime.NewMockClock(ctrl)
			fileRepo = mock_repository.NewMockFile(ctrl)
			fileClient = service.NewFile(service.FileParam{
				Validator:  validator,
				Identifier: identifier,
				Clock:      clock,
				FileRepo:   fileRepo,
			})
			p = service.SearchFileParam{
				Keyword:       "sa",
				TotalItems:    24,
				Page:          2,
				StatusIn:      []string{"available", "deleted"},
				VisibilityIn:  []string{"public"},
				ExtensionIn:   []string{"png", "jpg"},
				SizeGte:       1024,
				SizeLte:       2048,
				UploadDateGte: 1669638670000,
				UploadDateLte: 1669638670476,
				Sort:          "latest_upload",
			}
			searchParam = repository.SearchFileParam{
				Keyword:       "sa",
				Limit:         24,
				Offset:        24,
				StatusIn:      []string{"available", "deleted"},
				VisibilityIn:  []string{"public"},
				ExtensionIn:   []string{"png", "jpg"},
				SizeGte:       1024,
				SizeLte:       2048,
				UploadDateGte: 1669638670000,
				UploadDateLte: 1669638670476,
				Sort:          "latest_upload",
			}
			searchRes = &repository.SearchFileResult{
				Summary: repository.SearchFileSummary{
					TotalItems: 2,
				},
				Items: []repository.SearchFileItem{
					{
						Id:        "id-1",
						Name:      "name-1",
						Status:    "deleted",
						CreatedAt: currentTs,
					},
					{
						Id:        "id-2",
						Name:      "name-2",
						Status:    "available",
						CreatedAt: currentTs,
						UpdatedAt: &currentTs,
					},
				},
			}
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(fmt.Errorf("invalid data")).
					Times(1)

				res, err := fileClient.SearchFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("invalid data"))
			})
		})

		When("failed search file", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					SearchFile(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.SearchFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("there is no file", func() {
			It("should return empty result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				searchRes := &repository.SearchFileResult{
					Summary: repository.SearchFileSummary{
						TotalItems: 0,
					},
					Items: []repository.SearchFileItem{},
				}
				fileRepo.
					EXPECT().
					SearchFile(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				res, err := fileClient.SearchFile(ctx, p)

				Expect(res.Success.Code).To(Equal(int32(1000)))
				Expect(res.Success.Message).To(Equal("success search file"))
				Expect(res.Summary.Page).To(Equal(p.Page))
				Expect(res.Summary.TotalItems).To(Equal(int64(0)))
				Expect(len(res.Items)).To(Equal(0))
				Expect(err).To(BeNil())
			})
		})

		When("there is one file", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				searchRes := &repository.SearchFileResult{
					Summary: repository.SearchFileSummary{
						TotalItems: 1,
					},
					Items: []repository.SearchFileItem{
						{
							Id:        "id-1",
							Name:      "name-1",
							Status:    "deleted",
							CreatedAt: currentTs,
						},
					},
				}
				fileRepo.
					EXPECT().
					SearchFile(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				res, err := fileClient.SearchFile(ctx, p)

				Expect(res.Success.Code).To(Equal(int32(1000)))
				Expect(res.Success.Message).To(Equal("success search file"))
				Expect(res.Summary.Page).To(Equal(p.Page))
				Expect(res.Summary.TotalItems).To(Equal(int64(1)))
				Expect(len(res.Items)).To(Equal(1))
				Expect(err).To(BeNil())
			})
		})

		When("there are some files", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					SearchFile(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				res, err := fileClient.SearchFile(ctx, p)

				Expect(res.Success.Code).To(Equal(int32(1000)))
				Expect(res.Success.Message).To(Equal("success search file"))
				Expect(res.Summary.Page).To(Equal(p.Page))
				Expect(res.Summary.TotalItems).To(Equal(int64(2)))
				Expect(len(res.Items)).To(Equal(2))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("DeleteFileById function", Label("unit"), func() {
		var (
			ctx           context.Context
			currentTs     time.Time
			fileClient    service.File
			validator     *mock_validation.MockValidator
			identifier    *mock_identifier.MockIdentifier
			clock         *mock_datetime.MockClock
			slugger       *mock_slug.MockSlugger
			barrelRepo    *mock_repository.MockBarrel
			fileRepo      *mock_repository.MockFile
			storageRouter *mock_storage.MockRouter
			serializer    *mock_serialization.MockSerializer
			queuer        *mock_queueing.MockQueuer
			p             service.DeleteFileByIdParam
			r             *service.DeleteFileByIdResult
			findFileParam repository.FindFileParam
			findFileRes   *repository.FindFileResult
			updateParam   repository.UpdateFileParam
			updateRes     *repository.UpdateFileResult
			msgParam      *queue.DeleteFileMessage
			publishParam  queueing.PublishParam
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			validator = mock_validation.NewMockValidator(ctrl)
			identifier = mock_identifier.NewMockIdentifier(ctrl)
			clock = mock_datetime.NewMockClock(ctrl)
			slugger = mock_slug.NewMockSlugger(ctrl)
			barrelRepo = mock_repository.NewMockBarrel(ctrl)
			fileRepo = mock_repository.NewMockFile(ctrl)
			storageRouter = mock_storage.NewMockRouter(ctrl)
			serializer = mock_serialization.NewMockSerializer(ctrl)
			queuer = mock_queueing.NewMockQueuer(ctrl)
			fileClient = service.NewFile(service.FileParam{
				Validator:  validator,
				Identifier: identifier,
				Clock:      clock,
				Slugger:    slugger,
				BarrelRepo: barrelRepo,
				FileRepo:   fileRepo,
				Router:     storageRouter,
				Serializer: serializer,
				Pubsub:     queuer,
			})
			p = service.DeleteFileByIdParam{
				Id: "id",
			}
			findFileParam = repository.FindFileParam{
				Id: p.Id,
			}
			findFileRes = &repository.FindFileResult{
				Id:     findFileParam.Id,
				Status: "available",
				Locations: []repository.FindFileLocation{
					{
						Id: "l1",
						Barrel: repository.FindFileBarrel{
							Id: "b1",
						},
					},
				},
			}
			updateParam = repository.UpdateFileParam{
				Id:        findFileRes.Id,
				UpdatedAt: currentTs,
				Status:    typeconv.String("deleted"),
				DeletedAt: typeconv.Time(currentTs),
			}
			updateRes = &repository.UpdateFileResult{
				Id:        updateParam.Id,
				Status:    "deleted",
				UpdatedAt: currentTs,
				DeletedAt: typeconv.Time(currentTs),
			}
			msgParam = &queue.DeleteFileMessage{
				LocationId:  "l1",
				BarrelId:    "b1",
				FileId:      updateRes.Id,
				Status:      updateRes.Status,
				RequestedAt: updateRes.UpdatedAt.UnixMilli(),
			}
			publishParam = queueing.PublishParam{
				ExchangeName: "file_deletion",
				MessageBody:  []byte{},
			}
			r = &service.DeleteFileByIdResult{
				Success: system.Success{
					Code:    1000,
					Message: "success delete file",
				},
				RequestedAt: updateRes.UpdatedAt,
			}
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(fmt.Errorf("invalid data")).
					Times(1)

				res, err := fileClient.DeleteFileById(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("invalid data"))
			})
		})

		When("failed find file", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.DeleteFileById(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("file is not available", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(nil, repository.ErrNotFound).
					Times(1)

				res, err := fileClient.DeleteFileById(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1004)))
				Expect(err.Message).To(Equal("file is not available"))
			})
		})

		When("file is not delete able", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				findFileRes := &repository.FindFileResult{
					Status: "deleting",
				}
				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				res, err := fileClient.DeleteFileById(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1004)))
				Expect(err.Message).To(Equal("file is not available"))
			})
		})

		When("failed update file", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateFile(gomock.Eq(ctx), gomock.Eq(updateParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.DeleteFileById(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("failed marshal message", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateFile(gomock.Eq(ctx), gomock.Eq(updateParam)).
					Return(updateRes, nil).
					Times(1)

				serializer.
					EXPECT().
					Marshal(gomock.Eq(msgParam)).
					Return(nil, fmt.Errorf("i/o error")).
					Times(1)

				res, err := fileClient.DeleteFileById(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("i/o error"))
			})
		})

		When("failed publish message", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateFile(gomock.Eq(ctx), gomock.Eq(updateParam)).
					Return(updateRes, nil).
					Times(1)

				serializer.
					EXPECT().
					Marshal(gomock.Eq(msgParam)).
					Return([]byte{}, nil).
					Times(1)

				queuer.
					EXPECT().
					Publish(gomock.Eq(ctx), gomock.Eq(publishParam)).
					Return(fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.DeleteFileById(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("success delete file", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateFile(gomock.Eq(ctx), gomock.Eq(updateParam)).
					Return(updateRes, nil).
					Times(1)

				serializer.
					EXPECT().
					Marshal(gomock.Eq(msgParam)).
					Return([]byte{}, nil).
					Times(1)

				queuer.
					EXPECT().
					Publish(gomock.Eq(ctx), gomock.Eq(publishParam)).
					Return(nil).
					Times(1)

				res, err := fileClient.DeleteFileById(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("ProceedDeletion function", Label("unit"), func() {
		var (
			ctx            context.Context
			currentTs      time.Time
			fileClient     service.File
			validator      *mock_validation.MockValidator
			identifier     *mock_identifier.MockIdentifier
			clock          *mock_datetime.MockClock
			slugger        *mock_slug.MockSlugger
			barrelRepo     *mock_repository.MockBarrel
			fileRepo       *mock_repository.MockFile
			storageRouter  *mock_storage.MockRouter
			serializer     *mock_serialization.MockSerializer
			queuer         *mock_queueing.MockQueuer
			deleteStorage  *mock_storage.MockStorage
			p              service.ProceedDeletionParam
			r              *service.ProceedDeletionResult
			findFileParam  repository.FindFileParam
			findFileRes    *repository.FindFileResult
			deletingParam  repository.UpdateLocationByIdsParam
			createStgParam router.CreateStorageParam
			deleteObjParam storage.DeleteObjectParam
			deleteObjRes   *storage.DeleteObjectResult
			deletedParam   repository.UpdateLocationByIdsParam
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			validator = mock_validation.NewMockValidator(ctrl)
			identifier = mock_identifier.NewMockIdentifier(ctrl)
			clock = mock_datetime.NewMockClock(ctrl)
			slugger = mock_slug.NewMockSlugger(ctrl)
			barrelRepo = mock_repository.NewMockBarrel(ctrl)
			fileRepo = mock_repository.NewMockFile(ctrl)
			storageRouter = mock_storage.NewMockRouter(ctrl)
			serializer = mock_serialization.NewMockSerializer(ctrl)
			queuer = mock_queueing.NewMockQueuer(ctrl)
			deleteStorage = mock_storage.NewMockStorage(ctrl)
			fileClient = service.NewFile(service.FileParam{
				Validator:  validator,
				Identifier: identifier,
				Clock:      clock,
				Slugger:    slugger,
				BarrelRepo: barrelRepo,
				FileRepo:   fileRepo,
				Router:     storageRouter,
				Serializer: serializer,
				Pubsub:     queuer,
			})
			p = service.ProceedDeletionParam{
				LocationId: "l1",
			}
			r = &service.ProceedDeletionResult{
				Success: system.Success{
					Code:    1000,
					Message: "success delete file",
				},
				DeletedAt: currentTs,
			}
			findFileParam = repository.FindFileParam{
				LocationId: p.LocationId,
			}
			findFileRes = &repository.FindFileResult{
				Id:        findFileParam.Id,
				Status:    "deleted",
				DeletedAt: typeconv.Time(currentTs),
				Locations: []repository.FindFileLocation{
					{
						Id:     "l2",
						Status: "deleted",
						Barrel: repository.FindFileBarrel{
							Id:   "b2",
							Code: "b2",
						},
					},
					{
						Id:         "l1",
						Status:     "available",
						ExternalId: typeconv.String("e1"),
						Barrel: repository.FindFileBarrel{
							Id:   "b1",
							Code: "b1",
						},
					},
				},
			}
			deletingParam = repository.UpdateLocationByIdsParam{
				Ids:       []string{p.LocationId},
				UpdatedAt: currentTs,
				Status:    typeconv.String("deleting"),
			}
			createStgParam = router.CreateStorageParam{
				BarrelCode: "b1",
			}
			deleteObjParam = storage.DeleteObjectParam{
				ObjectId: "e1",
			}
			deleteObjRes = &storage.DeleteObjectResult{
				DeletedAt: currentTs,
			}
			deletedParam = repository.UpdateLocationByIdsParam{
				Ids:       []string{p.LocationId},
				UpdatedAt: currentTs,
				Status:    typeconv.String("deleted"),
				DeletedAt: typeconv.Time(currentTs),
			}
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(fmt.Errorf("invalid data")).
					Times(1)

				res, err := fileClient.ProceedDeletion(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("invalid data"))
			})
		})

		When("failed find file", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.ProceedDeletion(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("file is not available", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(nil, repository.ErrNotFound).
					Times(1)

				res, err := fileClient.ProceedDeletion(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1004)))
				Expect(err.Message).To(Equal("file is not available"))
			})
		})

		When("status is deleting", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				findFileRes := &repository.FindFileResult{
					Id:        findFileParam.Id,
					Status:    "deleted",
					DeletedAt: typeconv.Time(currentTs),
					Locations: []repository.FindFileLocation{
						{
							Id:     "l1",
							Status: "deleting",
							Barrel: repository.FindFileBarrel{
								Id: "b1",
							},
						},
					},
				}
				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				res, err := fileClient.ProceedDeletion(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1003)))
				Expect(err.Message).To(Equal("deletion is already proceeded"))
			})
		})

		When("failed create storage", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(nil, fmt.Errorf("config error")).
					Times(1)

				res, err := fileClient.ProceedDeletion(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("config error"))
			})
		})

		When("failed set deleting status", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(deleteStorage, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(deletingParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.ProceedDeletion(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("failed delete object", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(deleteStorage, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(deletingParam)).
					Return(nil, nil).
					Times(1)

				deleteStorage.
					EXPECT().
					DeleteObject(gomock.Eq(ctx), gomock.Eq(deleteObjParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.ProceedDeletion(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("failed set deleted status", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(deleteStorage, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(deletingParam)).
					Return(nil, nil).
					Times(1)

				deleteStorage.
					EXPECT().
					DeleteObject(gomock.Eq(ctx), gomock.Eq(deleteObjParam)).
					Return(deleteObjRes, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(deletedParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.ProceedDeletion(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("success proceed file deletion", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findFileParam)).
					Return(findFileRes, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(createStgParam)).
					Return(deleteStorage, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(deletingParam)).
					Return(nil, nil).
					Times(1)

				deleteStorage.
					EXPECT().
					DeleteObject(gomock.Eq(ctx), gomock.Eq(deleteObjParam)).
					Return(deleteObjRes, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(deletedParam)).
					Return(nil, nil).
					Times(1)

				res, err := fileClient.ProceedDeletion(ctx, p)

				Expect(err).To(BeNil())
				Expect(res).To(Equal(r))
			})
		})
	})

	Context("ScheduleReplication function", Label("unit"), func() {
		var (
			ctx           context.Context
			currentTs     time.Time
			fileClient    service.File
			validator     *mock_validation.MockValidator
			identifier    *mock_identifier.MockIdentifier
			clock         *mock_datetime.MockClock
			slugger       *mock_slug.MockSlugger
			barrelRepo    *mock_repository.MockBarrel
			fileRepo      *mock_repository.MockFile
			storageRouter *mock_storage.MockRouter
			serializer    *mock_serialization.MockSerializer
			p             service.ScheduleReplicationParam
			pubsub        *mock_queueing.MockPubsub
			r             *service.ScheduleReplicationResult
			searchParam   repository.SearchLocationParam
			searchRes     *repository.SearchLocationResult
			updateParam   repository.UpdateLocationByIdsParam
			publishParam  queueing.PublishParam
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			validator = mock_validation.NewMockValidator(ctrl)
			identifier = mock_identifier.NewMockIdentifier(ctrl)
			clock = mock_datetime.NewMockClock(ctrl)
			slugger = mock_slug.NewMockSlugger(ctrl)
			barrelRepo = mock_repository.NewMockBarrel(ctrl)
			fileRepo = mock_repository.NewMockFile(ctrl)
			storageRouter = mock_storage.NewMockRouter(ctrl)
			serializer = mock_serialization.NewMockSerializer(ctrl)
			pubsub = mock_queueing.NewMockPubsub(ctrl)
			fileClient = service.NewFile(service.FileParam{
				Validator:  validator,
				Identifier: identifier,
				Clock:      clock,
				Slugger:    slugger,
				BarrelRepo: barrelRepo,
				FileRepo:   fileRepo,
				Router:     storageRouter,
				Serializer: serializer,
				Pubsub:     pubsub,
			})

			p = service.ScheduleReplicationParam{
				MaxItems: 5,
			}
			searchParam = repository.SearchLocationParam{
				Limit:    p.MaxItems,
				Statuses: []string{"pending"},
			}
			searchRes = &repository.SearchLocationResult{
				Summary: repository.SearchLocationSummary{
					TotalItems: 3,
				},
				Items: []repository.SearchLocationItem{
					{
						Id:       "i1",
						FileId:   "f1",
						BarrelId: "b1",
						Priority: 2,
						Status:   "pending",
					},
					{
						Id:       "i2",
						FileId:   "f1",
						BarrelId: "b2",
						Priority: 2,
						Status:   "pending",
					},
					{
						Id:       "i3",
						FileId:   "f2",
						BarrelId: "b1",
						Priority: 2,
						Status:   "pending",
					},
				},
			}
			updateParam = repository.UpdateLocationByIdsParam{
				Ids:       []string{"i1", "i2", "i3"},
				Status:    typeconv.String("replicating"),
				UpdatedAt: currentTs,
			}
			publishParam = queueing.PublishParam{
				ExchangeName: "file_replication",
				MessageBody:  []byte{},
			}
			r = &service.ScheduleReplicationResult{
				Success: system.Success{
					Code:    1000,
					Message: "success schedule replication",
				},
				TotalItems:  3,
				ScheduledAt: &currentTs,
			}
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(fmt.Errorf("invalid data")).
					Times(1)

				res, err := fileClient.ScheduleReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("invalid data"))
			})
		})

		When("failed search location", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					SearchLocation(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.ScheduleReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("there is no pending replication", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				searchRes := &repository.SearchLocationResult{
					Summary: repository.SearchLocationSummary{
						TotalItems: 0,
					},
					Items: []repository.SearchLocationItem{},
				}
				fileRepo.
					EXPECT().
					SearchLocation(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				res, err := fileClient.ScheduleReplication(ctx, p)

				r := &service.ScheduleReplicationResult{
					Success: system.Success{
						Code:    1000,
						Message: "there is no pending replication",
					},
					TotalItems: 0,
				}
				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("failed marshall message", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					SearchLocation(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				serializer.
					EXPECT().
					Marshal(gomock.Any()).
					Return(nil, fmt.Errorf("marshall error")).
					Times(1)

				res, err := fileClient.ScheduleReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("marshall error"))
			})
		})

		When("failed update location", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					SearchLocation(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				serializer.
					EXPECT().
					Marshal(gomock.Any()).
					Return([]byte{}, nil).
					Times(3)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(updateParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.ScheduleReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("failed publish message", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					SearchLocation(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				serializer.
					EXPECT().
					Marshal(gomock.Any()).
					Return([]byte{}, nil).
					Times(3)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(updateParam)).
					Return(nil, nil).
					Times(1)

				pubsub.
					EXPECT().
					Publish(gomock.Eq(ctx), gomock.Eq(publishParam)).
					Return(fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.ScheduleReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("success schedule replication", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					SearchLocation(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				serializer.
					EXPECT().
					Marshal(gomock.Any()).
					Return([]byte{}, nil).
					Times(3)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(updateParam)).
					Return(nil, nil).
					Times(1)

				pubsub.
					EXPECT().
					Publish(gomock.Eq(ctx), gomock.Eq(publishParam)).
					Return(nil).
					Times(3)

				res, err := fileClient.ScheduleReplication(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("ProceedReplication function", Label("unit"), func() {
		var (
			ctx               context.Context
			currentTs         time.Time
			fileClient        service.File
			validator         *mock_validation.MockValidator
			identifier        *mock_identifier.MockIdentifier
			clock             *mock_datetime.MockClock
			slugger           *mock_slug.MockSlugger
			barrelRepo        *mock_repository.MockBarrel
			fileRepo          *mock_repository.MockFile
			storageRouter     *mock_storage.MockRouter
			primaryStorage    *mock_storage.MockStorage
			replicaStorage    *mock_storage.MockStorage
			serializer        *mock_serialization.MockSerializer
			pubsub            *mock_queueing.MockPubsub
			p                 service.ProceedReplicationParam
			findParam         repository.FindFileParam
			findRes           *repository.FindFileResult
			primParam         router.CreateStorageParam
			replParam         router.CreateStorageParam
			retrParam         storage.RetrieveObjectParam
			retrRes           *storage.RetrieveObjectResult
			uploadRes         *storage.UploadObjectResult
			updateUploadParam repository.UpdateLocationByIdsParam
			updateUploadRes   *repository.UpdateLocationByIdsResult
			updateParam       repository.UpdateLocationByIdsParam
			updateRes         *repository.UpdateLocationByIdsResult
			r                 *service.ProceedReplicationResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			validator = mock_validation.NewMockValidator(ctrl)
			identifier = mock_identifier.NewMockIdentifier(ctrl)
			clock = mock_datetime.NewMockClock(ctrl)
			slugger = mock_slug.NewMockSlugger(ctrl)
			barrelRepo = mock_repository.NewMockBarrel(ctrl)
			fileRepo = mock_repository.NewMockFile(ctrl)
			storageRouter = mock_storage.NewMockRouter(ctrl)
			primaryStorage = mock_storage.NewMockStorage(ctrl)
			replicaStorage = mock_storage.NewMockStorage(ctrl)
			serializer = mock_serialization.NewMockSerializer(ctrl)
			pubsub = mock_queueing.NewMockPubsub(ctrl)
			fileClient = service.NewFile(service.FileParam{
				Validator:  validator,
				Identifier: identifier,
				Clock:      clock,
				Slugger:    slugger,
				BarrelRepo: barrelRepo,
				FileRepo:   fileRepo,
				Router:     storageRouter,
				Serializer: serializer,
				Pubsub:     pubsub,
			})

			p = service.ProceedReplicationParam{
				LocationId: "loc2",
			}
			findParam = repository.FindFileParam{
				LocationId: p.LocationId,
			}
			findRes = &repository.FindFileResult{
				Id:        "id",
				Name:      "dolphin 22",
				Mimetype:  "image/jpeg",
				Extension: "jpg",
				Status:    "available",
				Locations: []repository.FindFileLocation{
					{
						Id:       "loc2",
						Priority: 2,
						Status:   "replicating",
						Barrel: repository.FindFileBarrel{
							Id:   "b2",
							Code: "b2",
						},
					},
					{
						Id:         "loc1",
						Priority:   1,
						Status:     "available",
						ExternalId: typeconv.String("e1"),
						Barrel: repository.FindFileBarrel{
							Id:   "b1",
							Code: "b1",
						},
					},
				},
			}
			primParam = router.CreateStorageParam{
				BarrelCode: "b1",
			}
			replParam = router.CreateStorageParam{
				BarrelCode: "b2",
			}
			retrParam = storage.RetrieveObjectParam{
				ObjectId: "e1",
			}
			retrRes = &storage.RetrieveObjectResult{
				Data:        nil,
				RetrievedAt: currentTs,
			}
			uploadRes = &storage.UploadObjectResult{
				ObjectId:   "e2",
				UploadedAt: currentTs,
			}
			updateUploadParam = repository.UpdateLocationByIdsParam{
				Ids:       []string{"loc2"},
				Status:    typeconv.String("uploading"),
				UpdatedAt: currentTs,
			}
			updateUploadRes = &repository.UpdateLocationByIdsResult{
				TotalUpdated: 1,
			}
			updateParam = repository.UpdateLocationByIdsParam{
				Ids:        []string{"loc2"},
				Status:     typeconv.String("available"),
				ExternalId: typeconv.String("e2"),
				UploadedAt: typeconv.Time(currentTs),
				UpdatedAt:  currentTs,
			}
			updateRes = &repository.UpdateLocationByIdsResult{
				TotalUpdated: 1,
			}
			r = &service.ProceedReplicationResult{
				Success: system.Success{
					Code:    1000,
					Message: "success replicate file",
				},
				LocationId: typeconv.String("loc2"),
				BarrelId:   typeconv.String("b2"),
				ExternalId: typeconv.String("e2"),
				UploadedAt: typeconv.Time(currentTs),
			}
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(fmt.Errorf("invalid data")).
					Times(1)

				res, err := fileClient.ProceedReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("invalid data"))
			})
		})

		When("failed find file", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.ProceedReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("file is not available", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(nil, repository.ErrNotFound).
					Times(1)

				res, err := fileClient.ProceedReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1004)))
				Expect(err.Message).To(Equal("file is not available"))
			})
		})

		When("status is not replicating", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				findRes := &repository.FindFileResult{
					Status: "available",
					Locations: []repository.FindFileLocation{
						{
							Priority: 2,
							Status:   "available",
						},
						{
							Priority: 1,
							Status:   "available",
						},
					},
				}
				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				res, err := fileClient.ProceedReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1003)))
				Expect(err.Message).To(Equal("replication is already proceeded"))
			})
		})

		When("failed update status uploading", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(updateUploadParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.ProceedReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("failed create primary storage", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(updateUploadParam)).
					Return(updateUploadRes, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(primParam)).
					Return(nil, router.ErrUnsupported).
					Times(1)

				res, err := fileClient.ProceedReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("unsupported storage"))
			})
		})

		When("failed create replica storage", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(updateUploadParam)).
					Return(updateUploadRes, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(primParam)).
					Return(primaryStorage, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(replParam)).
					Return(nil, router.ErrUnsupported).
					Times(1)

				res, err := fileClient.ProceedReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("unsupported storage"))
			})
		})

		When("failed retrieve object from primary storage", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(updateUploadParam)).
					Return(updateUploadRes, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(primParam)).
					Return(primaryStorage, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(replParam)).
					Return(replicaStorage, nil).
					Times(1)

				primaryStorage.
					EXPECT().
					RetrieveObject(gomock.Eq(ctx), gomock.Eq(retrParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.ProceedReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("failed upload object to replica storage", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(updateUploadParam)).
					Return(updateUploadRes, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(primParam)).
					Return(primaryStorage, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(replParam)).
					Return(replicaStorage, nil).
					Times(1)

				primaryStorage.
					EXPECT().
					RetrieveObject(gomock.Eq(ctx), gomock.Eq(retrParam)).
					Return(retrRes, nil).
					Times(1)

				replicaStorage.
					EXPECT().
					UploadObject(gomock.Eq(ctx), gomock.Any()).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.ProceedReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("failed update location", func() {
			It("should return error", func() {
				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(updateUploadParam)).
					Return(updateUploadRes, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(primParam)).
					Return(primaryStorage, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(replParam)).
					Return(replicaStorage, nil).
					Times(1)

				primaryStorage.
					EXPECT().
					RetrieveObject(gomock.Eq(ctx), gomock.Eq(retrParam)).
					Return(retrRes, nil).
					Times(1)

				replicaStorage.
					EXPECT().
					UploadObject(gomock.Eq(ctx), gomock.Any()).
					Return(uploadRes, nil).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(updateParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := fileClient.ProceedReplication(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("success replicate file", func() {
			It("should return result", func() {
				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				fileRepo.
					EXPECT().
					FindFile(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(updateUploadParam)).
					Return(updateUploadRes, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(primParam)).
					Return(primaryStorage, nil).
					Times(1)

				storageRouter.
					EXPECT().
					CreateStorage(gomock.Eq(ctx), gomock.Eq(replParam)).
					Return(replicaStorage, nil).
					Times(1)

				primaryStorage.
					EXPECT().
					RetrieveObject(gomock.Eq(ctx), gomock.Eq(retrParam)).
					Return(retrRes, nil).
					Times(1)

				replicaStorage.
					EXPECT().
					UploadObject(gomock.Eq(ctx), gomock.Any()).
					Return(uploadRes, nil).
					Times(1)

				fileRepo.
					EXPECT().
					UpdateLocationByIds(gomock.Eq(ctx), gomock.Eq(updateParam)).
					Return(updateRes, nil).
					Times(1)

				res, err := fileClient.ProceedReplication(ctx, p)

				Expect(err).To(BeNil())
				Expect(res).To(Equal(r))
			})
		})
	})
})
