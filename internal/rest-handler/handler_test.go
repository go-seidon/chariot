package rest_handler_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRestHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rest Handler Package")
}
