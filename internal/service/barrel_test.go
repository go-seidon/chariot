package service_test

import (
	"context"
	"fmt"
	"time"

	"github.com/go-seidon/chariot/internal/repository"
	mock_repository "github.com/go-seidon/chariot/internal/repository/mock"
	"github.com/go-seidon/chariot/internal/service"
	mock_datetime "github.com/go-seidon/provider/datetime/mock"
	mock_identifier "github.com/go-seidon/provider/identity/mock"
	"github.com/go-seidon/provider/system"
	mock_validation "github.com/go-seidon/provider/validation/mock"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Barrel Package", func() {
	Context("CreateBarrel function", Label("unit"), func() {
		var (
			ctx           context.Context
			currentTs     time.Time
			barrelService service.Barrel
			p             service.CreateBarrelParam
			validator     *mock_validation.MockValidator
			identifier    *mock_identifier.MockIdentifier
			clock         *mock_datetime.MockClock
			barrelRepo    *mock_repository.MockBarrel
			createParam   repository.CreateBarrelParam
			createRes     *repository.CreateBarrelResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			validator = mock_validation.NewMockValidator(ctrl)
			identifier = mock_identifier.NewMockIdentifier(ctrl)
			clock = mock_datetime.NewMockClock(ctrl)
			barrelRepo = mock_repository.NewMockBarrel(ctrl)
			barrelService = service.NewBarrel(service.BarrelParam{
				Validator:  validator,
				Identifier: identifier,
				Clock:      clock,
				BarrelRepo: barrelRepo,
			})
			p = service.CreateBarrelParam{
				Code:     "code",
				Name:     "name",
				Provider: "goseidon_hippo",
				Status:   "active",
			}
			createParam = repository.CreateBarrelParam{
				Id:        "id",
				Code:      p.Code,
				Name:      p.Name,
				Provider:  p.Provider,
				Status:    p.Status,
				CreatedAt: currentTs,
			}
			createRes = &repository.CreateBarrelResult{
				Id:        "id",
				Code:      p.Code,
				Name:      p.Name,
				Provider:  p.Provider,
				Status:    p.Status,
				CreatedAt: currentTs,
			}
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(fmt.Errorf("invalid data")).
					Times(1)

				res, err := barrelService.CreateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("invalid data"))
			})
		})

		When("failed generate id", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("", fmt.Errorf("generate error")).
					Times(1)

				res, err := barrelService.CreateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("generate error"))
			})
		})

		When("failed create barrel", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("id", nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				barrelRepo.
					EXPECT().
					CreateBarrel(gomock.Eq(ctx), gomock.Eq(createParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := barrelService.CreateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("barrel is already exists", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("id", nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				barrelRepo.
					EXPECT().
					CreateBarrel(gomock.Eq(ctx), gomock.Eq(createParam)).
					Return(nil, repository.ErrExists).
					Times(1)

				res, err := barrelService.CreateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("barrel is already exists"))
			})
		})

		When("success create barrel", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				identifier.
					EXPECT().
					GenerateId().
					Return("id", nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				barrelRepo.
					EXPECT().
					CreateBarrel(gomock.Eq(ctx), gomock.Eq(createParam)).
					Return(createRes, nil).
					Times(1)

				res, err := barrelService.CreateBarrel(ctx, p)

				Expect(err).To(BeNil())
				Expect(res.Success.Code).To(Equal(int32(1000)))
				Expect(res.Success.Message).To(Equal("success create barrel"))
				Expect(res.Id).To(Equal("id"))
				Expect(res.Code).To(Equal("code"))
				Expect(res.Name).To(Equal("name"))
				Expect(res.Status).To(Equal("active"))
				Expect(res.Provider).To(Equal("goseidon_hippo"))
				Expect(res.CreatedAt).To(Equal(currentTs))
			})
		})
	})

	Context("FindBarrelById function", Label("unit"), func() {

		var (
			ctx           context.Context
			currentTs     time.Time
			barrelService service.Barrel
			param         service.FindBarrelByIdParam
			result        *service.FindBarrelByIdResult
			validator     *mock_validation.MockValidator
			identifier    *mock_identifier.MockIdentifier
			clock         *mock_datetime.MockClock
			barrelRepo    *mock_repository.MockBarrel
			findParam     repository.FindBarrelParam
			findRes       *repository.FindBarrelResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			validator = mock_validation.NewMockValidator(ctrl)
			identifier = mock_identifier.NewMockIdentifier(ctrl)
			clock = mock_datetime.NewMockClock(ctrl)
			barrelRepo = mock_repository.NewMockBarrel(ctrl)
			barrelService = service.NewBarrel(service.BarrelParam{
				Validator:  validator,
				Identifier: identifier,
				Clock:      clock,
				BarrelRepo: barrelRepo,
			})

			param = service.FindBarrelByIdParam{
				Id: "id",
			}
			findParam = repository.FindBarrelParam{
				Id: param.Id,
			}
			findRes = &repository.FindBarrelResult{
				Id:        "id",
				Code:      "code",
				Name:      "name",
				Provider:  "goseidon_hippo",
				Status:    "active",
				CreatedAt: currentTs,
			}
			result = &service.FindBarrelByIdResult{
				Success: system.Success{
					Code:    1000,
					Message: "success find barrel",
				},
				Id:        findRes.Id,
				Code:      findRes.Code,
				Name:      findRes.Name,
				Provider:  findRes.Provider,
				Status:    findRes.Status,
				CreatedAt: findRes.CreatedAt,
				UpdatedAt: findRes.UpdatedAt,
			}
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(param)).
					Return(fmt.Errorf("invalid data")).
					Times(1)

				res, err := barrelService.FindBarrelById(ctx, param)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("invalid data"))
			})
		})

		When("failed find barrel", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(param)).
					Return(nil).
					Times(1)

				barrelRepo.
					EXPECT().
					FindBarrel(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := barrelService.FindBarrelById(ctx, param)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("barrel is not available", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(param)).
					Return(nil).
					Times(1)

				barrelRepo.
					EXPECT().
					FindBarrel(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(nil, repository.ErrNotFound).
					Times(1)

				res, err := barrelService.FindBarrelById(ctx, param)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1004)))
				Expect(err.Message).To(Equal("barrel is not available"))
			})
		})

		When("barrel is available", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(param)).
					Return(nil).
					Times(1)

				barrelRepo.
					EXPECT().
					FindBarrel(gomock.Eq(ctx), gomock.Eq(findParam)).
					Return(findRes, nil).
					Times(1)

				res, err := barrelService.FindBarrelById(ctx, param)

				Expect(res).To(Equal(result))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("UpdateBarrelById function", Label("unit"), func() {
		var (
			ctx          context.Context
			currentTs    time.Time
			barrelClient service.Barrel
			p            service.UpdateBarrelByIdParam
			validator    *mock_validation.MockValidator
			identifier   *mock_identifier.MockIdentifier
			clock        *mock_datetime.MockClock
			barrelRepo   *mock_repository.MockBarrel
			updateParam  repository.UpdateBarrelParam
			updateRes    *repository.UpdateBarrelResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			validator = mock_validation.NewMockValidator(ctrl)
			identifier = mock_identifier.NewMockIdentifier(ctrl)
			clock = mock_datetime.NewMockClock(ctrl)
			barrelRepo = mock_repository.NewMockBarrel(ctrl)
			barrelClient = service.NewBarrel(service.BarrelParam{
				Validator:  validator,
				Identifier: identifier,
				Clock:      clock,
				BarrelRepo: barrelRepo,
			})
			p = service.UpdateBarrelByIdParam{
				Id:       "id",
				Code:     "code",
				Name:     "name",
				Provider: "goseidon_hippo",
				Status:   "active",
			}
			updateParam = repository.UpdateBarrelParam{
				Id:        "id",
				Code:      p.Code,
				Name:      p.Name,
				Provider:  p.Provider,
				Status:    p.Status,
				UpdatedAt: currentTs,
			}
			updateRes = &repository.UpdateBarrelResult{
				Id:        "id",
				Code:      p.Code,
				Name:      p.Name,
				Provider:  p.Provider,
				Status:    p.Status,
				CreatedAt: currentTs,
				UpdatedAt: currentTs,
			}
		})

		When("there is invalid data", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(fmt.Errorf("invalid data")).
					Times(1)

				res, err := barrelClient.UpdateBarrelById(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("invalid data"))
			})
		})

		When("failed update barrel", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				barrelRepo.
					EXPECT().
					UpdateBarrel(gomock.Eq(ctx), gomock.Eq(updateParam)).
					Return(nil, fmt.Errorf("network error")).
					Times(1)

				res, err := barrelClient.UpdateBarrelById(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("barrel is not available", func() {
			It("should return error", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				barrelRepo.
					EXPECT().
					UpdateBarrel(gomock.Eq(ctx), gomock.Eq(updateParam)).
					Return(nil, repository.ErrNotFound).
					Times(1)

				res, err := barrelClient.UpdateBarrelById(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1004)))
				Expect(err.Message).To(Equal("barrel is not available"))
			})
		})

		When("success update barrel", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				clock.
					EXPECT().
					Now().
					Return(currentTs).
					Times(1)

				barrelRepo.
					EXPECT().
					UpdateBarrel(gomock.Eq(ctx), gomock.Eq(updateParam)).
					Return(updateRes, nil).
					Times(1)

				res, err := barrelClient.UpdateBarrelById(ctx, p)

				Expect(res.Success.Code).To(Equal(int32(1000)))
				Expect(res.Success.Message).To(Equal("success update barrel"))
				Expect(res.Id).To(Equal("id"))
				Expect(res.Name).To(Equal("name"))
				Expect(res.Status).To(Equal("active"))
				Expect(res.Provider).To(Equal("goseidon_hippo"))
				Expect(res.CreatedAt).To(Equal(currentTs))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("SearchBarrel function", Label("unit"), func() {
		var (
			ctx          context.Context
			currentTs    time.Time
			barrelClient service.Barrel
			p            service.SearchBarrelParam
			validator    *mock_validation.MockValidator
			identifier   *mock_identifier.MockIdentifier
			clock        *mock_datetime.MockClock
			barrelRepo   *mock_repository.MockBarrel
			searchParam  repository.SearchBarrelParam
			searchRes    *repository.SearchBarrelResult
		)

		BeforeEach(func() {
			ctx = context.Background()
			currentTs = time.Now().UTC()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			validator = mock_validation.NewMockValidator(ctrl)
			identifier = mock_identifier.NewMockIdentifier(ctrl)
			clock = mock_datetime.NewMockClock(ctrl)
			barrelRepo = mock_repository.NewMockBarrel(ctrl)
			barrelClient = service.NewBarrel(service.BarrelParam{
				Validator:  validator,
				Identifier: identifier,
				Clock:      clock,
				BarrelRepo: barrelRepo,
			})
			p = service.SearchBarrelParam{
				Keyword:    "goseidon",
				TotalItems: 24,
				Page:       2,
				Statuses:   []string{"active", "inactive"},
				Providers:  []string{"goseidon_hippo"},
			}
			searchParam = repository.SearchBarrelParam{
				Limit:     24,
				Offset:    24,
				Keyword:   "goseidon",
				Statuses:  []string{"active", "inactive"},
				Providers: []string{"goseidon_hippo"},
			}
			searchRes = &repository.SearchBarrelResult{
				Summary: repository.SearchBarrelSummary{
					TotalItems: 2,
				},
				Items: []repository.SearchBarrelItem{
					{
						Id:        "id-1",
						Code:      "code-1",
						Name:      "name-1",
						Provider:  "goseidon_hippo",
						Status:    "inactive",
						CreatedAt: currentTs,
					},
					{
						Id:        "id-2",
						Code:      "code-2",
						Name:      "name-2",
						Provider:  "goseidon_hippo",
						Status:    "active",
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

				res, err := barrelClient.SearchBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1002)))
				Expect(err.Message).To(Equal("invalid data"))
			})
		})

		When("failed search barrel", func() {
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

				res, err := barrelClient.SearchBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err.Code).To(Equal(int32(1001)))
				Expect(err.Message).To(Equal("network error"))
			})
		})

		When("there is no barrel", func() {
			It("should return empty result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				searchRes := &repository.SearchBarrelResult{
					Summary: repository.SearchBarrelSummary{
						TotalItems: 0,
					},
					Items: []repository.SearchBarrelItem{},
				}
				barrelRepo.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				res, err := barrelClient.SearchBarrel(ctx, p)

				Expect(res.Success.Code).To(Equal(int32(1000)))
				Expect(res.Success.Message).To(Equal("success search barrel"))
				Expect(res.Summary.Page).To(Equal(p.Page))
				Expect(res.Summary.TotalItems).To(Equal(int64(0)))
				Expect(len(res.Items)).To(Equal(0))
				Expect(err).To(BeNil())
			})
		})

		When("there is one barrel", func() {
			It("should return result", func() {
				validator.
					EXPECT().
					Validate(gomock.Eq(p)).
					Return(nil).
					Times(1)

				searchRes := &repository.SearchBarrelResult{
					Summary: repository.SearchBarrelSummary{
						TotalItems: 1,
					},
					Items: []repository.SearchBarrelItem{
						{
							Id:        "id-1",
							Code:      "code-1",
							Name:      "name-1",
							Provider:  "goseidon_hippo",
							Status:    "inactive",
							CreatedAt: currentTs,
						},
					},
				}
				barrelRepo.
					EXPECT().
					SearchBarrel(gomock.Eq(ctx), gomock.Eq(searchParam)).
					Return(searchRes, nil).
					Times(1)

				res, err := barrelClient.SearchBarrel(ctx, p)

				Expect(res.Success.Code).To(Equal(int32(1000)))
				Expect(res.Success.Message).To(Equal("success search barrel"))
				Expect(res.Summary.Page).To(Equal(p.Page))
				Expect(res.Summary.TotalItems).To(Equal(int64(1)))
				Expect(len(res.Items)).To(Equal(1))
				Expect(err).To(BeNil())
			})
		})

		When("there are some barrels", func() {
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

				res, err := barrelClient.SearchBarrel(ctx, p)

				Expect(res.Success.Code).To(Equal(int32(1000)))
				Expect(res.Success.Message).To(Equal("success search barrel"))
				Expect(res.Summary.Page).To(Equal(p.Page))
				Expect(res.Summary.TotalItems).To(Equal(int64(2)))
				Expect(len(res.Items)).To(Equal(2))
				Expect(err).To(BeNil())
			})
		})
	})

})
