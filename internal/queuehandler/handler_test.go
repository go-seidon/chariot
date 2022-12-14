package queuehandler_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestQueueHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Queue Handler Package")
}
