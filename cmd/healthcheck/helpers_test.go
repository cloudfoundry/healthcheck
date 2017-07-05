// +build !windows

package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func createPortHealthCheck(args []string, port string) *gexec.Session {
	args = append([]string{"-port", port, "-timeout", "100ms"}, args...)
	session, err := gexec.Start(exec.Command(healthCheck, args...), GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

func createHTTPHealthCheck(args []string, port string) *gexec.Session {
	args = append([]string{"-uri", "/api/_ping", "-port", port, "-timeout", "100ms"}, args...)
	session, err := gexec.Start(exec.Command(healthCheck, args...), GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}
