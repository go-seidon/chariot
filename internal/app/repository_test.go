package app_test

import (
	"fmt"

	"github.com/go-seidon/chariot/internal/app"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Repository Package", func() {

	Context("NewDefaultRepository function", Label("unit"), func() {
		When("config is not specified", func() {
			It("should return error", func() {
				res, err := app.NewDefaultRepository(nil)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid config")))
			})
		})

		When("db provider is not valid", func() {
			It("should return error", func() {
				res, err := app.NewDefaultRepository(&app.Config{
					RepositoryProvider: "invalid",
				})

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid repository provider")))
			})
		})

		Context("mysql repository", func() {
			When("success create repository", func() {
				It("should return result", func() {
					res, err := app.NewDefaultRepository(&app.Config{
						RepositoryProvider: "mysql",
					})

					Expect(res).ToNot(BeNil())
					Expect(err).To(BeNil())
				})
			})
		})
	})

})
