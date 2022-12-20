package app_test

import (
	"github.com/go-seidon/chariot/internal/app"
	mock_repository "github.com/go-seidon/chariot/internal/repository/mock"
	mock_logging "github.com/go-seidon/provider/logging/mock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Healthcheck Package", func() {

	Context("NewDefaultHealthCheck function", Label("unit"), func() {
		var (
			logger     *mock_logging.MockLogger
			repository *mock_repository.MockRepository
		)

		BeforeEach(func() {
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			logger = mock_logging.NewMockLogger(ctrl)
			repository = mock_repository.NewMockRepository(ctrl)
		})

		When("success create default healthcheck", func() {
			It("should return result", func() {
				res, err := app.NewDefaultHealthCheck(logger, repository)

				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("failed create default healthcheck", func() {
			It("should return error", func() {
				res, err := app.NewDefaultHealthCheck(nil, nil)

				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})
	})

})
