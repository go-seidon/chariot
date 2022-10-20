package app_test

import (
	"fmt"

	"github.com/go-seidon/chariot/internal/app"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logging Package", func() {

	Context("NewDefaultLog function", Label("unit"), func() {
		var (
			config *app.Config
		)

		BeforeEach(func() {
			config = &app.Config{
				AppDebug: true,
				AppEnv:   "local",
			}
		})

		When("config is not specified", func() {
			It("should return error", func() {
				res, err := app.NewDefaultLog(nil, "app-name")

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid config")))
			})
		})

		When("app name is not specified", func() {
			It("should return error", func() {
				res, err := app.NewDefaultLog(config, "")

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid app name")))
			})
		})

		When("success create default log", func() {
			It("should return result", func() {
				res, err := app.NewDefaultLog(config, "app-name")

				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})
	})

})
