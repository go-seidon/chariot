package app_test

import (
	"fmt"

	"github.com/go-seidon/chariot/internal/app"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Queueing Package", func() {

	Context("NewDefaultQueueing function", Label("unit"), func() {
		When("config is not specified", func() {
			It("should return error", func() {
				res, err := app.NewDefaultQueueing(nil)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid config")))
			})
		})

		When("provider is not supported", func() {
			It("should return error", func() {
				res, err := app.NewDefaultQueueing(&app.Config{
					QueueProvider: "unsupported",
				})

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid queue provider")))
			})
		})

		When("success create queue", func() {
			It("should return result", func() {
				res, err := app.NewDefaultQueueing(&app.Config{
					QueueProvider: "rabbitmq",
				})

				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})
	})
})
