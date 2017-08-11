package main_test

import (
	"net"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os"
	"io/ioutil"
)

var _ = Describe("Main", func() {
	var (
		session *gexec.Session
		command *exec.Cmd
		err     error
	)

	BeforeEach(func() {
		pluginsDir, err := ioutil.TempDir(os.TempDir(), "plugin-path")
		Expect(err).ToNot(HaveOccurred())

		os.MkdirAll(pluginsDir, os.ModePerm)
		Expect(err).ToNot(HaveOccurred())

		command = exec.Command(driverPath, "--listenAddr", "0.0.0.0:50052", "--pluginsPath", pluginsDir)
	})

	JustBeforeEach(func() {
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		session.Kill().Wait()
	})

  Context("with a driver path", func() {
    It("listens on tcp/50052 by default", func() {
      EventuallyWithOffset(1, func() error {
        _, err := net.Dial("tcp", "127.0.0.1:50052")
        return err
      }, 5).ShouldNot(HaveOccurred())
    })

	})
})
