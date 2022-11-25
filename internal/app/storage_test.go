package app_test

import (
	"fmt"

	"github.com/go-seidon/chariot/internal/app"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Storage Package", func() {

	Context("NewDefaultStorageRouter function", func() {
		var (
			p app.StorageRouterParam
		)

		BeforeEach(func() {
			p = app.StorageRouterParam{
				Config: &app.Config{
					Barrels: map[string]app.BarrelConfig{
						"hippo1": {
							Provider: "goseidon_hippo",
						},
						"hippo2": {
							Provider: "goseidon_hippo",
						},
					},
				},
			}
		})

		When("config is not specified", func() {
			It("should return error", func() {
				p.Config = nil
				res, err := app.NewDefaultStorageRouter(p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid config")))
			})
		})

		When("barrels are empty", func() {
			It("should return result", func() {
				p.Config.Barrels = map[string]app.BarrelConfig{}
				res, err := app.NewDefaultStorageRouter(p)

				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("provider is not supported", func() {
			It("should return error", func() {
				p.Config.Barrels = map[string]app.BarrelConfig{
					"hippo1": {
						Provider: "unsupported",
					},
				}
				res, err := app.NewDefaultStorageRouter(p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid barrel provider")))
			})
		})

		When("there is one barrel", func() {
			It("should return result", func() {
				p.Config.Barrels = map[string]app.BarrelConfig{
					"hippo1": {
						Provider: "goseidon_hippo",
					},
				}
				res, err := app.NewDefaultStorageRouter(p)

				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("there are multiple barrel", func() {
			It("should return result", func() {
				res, err := app.NewDefaultStorageRouter(p)

				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})
	})
})
