package main_test

import (
  "net"
  "os/exec"

  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  "github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
  var (
    session *gexec.Session
    command *exec.Cmd
    err     error
  )

  BeforeEach(func() {
    command = exec.Command(driverPath)
  })

  JustBeforeEach(func() {
    session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
    Expect(err).ToNot(HaveOccurred())
  })

  AfterEach(func() {
    session.Kill().Wait()
  })

  Context("with a driver path", func() {
    It("listens on tcp/7589 by default", func() {
      EventuallyWithOffset(1, func() error {
        _, err := net.Dial("tcp", "0.0.0.0:50051")
        return err
      }, 5).ShouldNot(HaveOccurred())
    })

  })
})