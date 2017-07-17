package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestNodeservice(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nodeservice Suite")
}
