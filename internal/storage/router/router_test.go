package router_test

import (
	"context"
	"testing"

	"github.com/go-seidon/chariot/internal/storage"
	mock_storage "github.com/go-seidon/chariot/internal/storage/mock"
	"github.com/go-seidon/chariot/internal/storage/router"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRouter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Router Package")
}

var _ = Describe("Storage Router", func() {

	Context("CreateStorage function", Label("unit"), func() {
		var (
			r   router.Router
			ctx context.Context
			p   router.CreateStorageParam
		)

		BeforeEach(func() {
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			st := mock_storage.NewMockStorage(ctrl)
			r = router.NewRouter(router.RouterParam{
				Barrels: map[string]storage.Storage{
					"s31":    st,
					"hippo1": st,
				},
			})
			ctx = context.Background()
			p = router.CreateStorageParam{
				BarrelCode: "hippo1",
			}
		})

		When("storage is not available", func() {
			It("should return error", func() {
				r := router.NewRouter(router.RouterParam{})
				res, err := r.CreateStorage(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(router.ErrUnsupported))
			})
		})

		When("storage is available", func() {
			It("should return result", func() {
				res, err := r.CreateStorage(ctx, p)

				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})
	})

})
