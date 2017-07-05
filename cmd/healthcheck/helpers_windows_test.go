package main_test

import (
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func createHTTPHealthCheck(args []string, port string) *gexec.Session {
	command := exec.Command(healthCheck, "-uri", "/api/_ping", "-port", "8080", "-timeout", "100ms")
	command.Env = append(
		os.Environ(),
		fmt.Sprintf(`CF_INSTANCE_PORTS=[{"external":%s,"internal":%s}]`, port, "8080"),
	)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

func createPortHealthCheck(args []string, port string) *gexec.Session {
	command := exec.Command(healthCheck, "-port", "8080", "-timeout", "100ms")
	command.Env = append(
		os.Environ(),
		fmt.Sprintf(`CF_INSTANCE_PORTS=[{"external":%s,"internal":%s}]`, port, "8080"),
	)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}
