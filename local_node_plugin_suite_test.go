package local_node_plugin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLocalNodePlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LocalNodePlugin Suite")
}
