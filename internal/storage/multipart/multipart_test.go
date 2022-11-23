package multipart_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMultipart(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multipart Package")
}
