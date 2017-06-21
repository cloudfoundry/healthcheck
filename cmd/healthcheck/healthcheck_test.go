// +build !windows

package main_test

import (
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("HealthCheck", func() {
	var (
		server     *ghttp.Server
		serverAddr string
		port       string
		args       []string
	)

	itExitsWithCode := func(healthCheck func() *gexec.Session, code int, reason string) {
		It("exits with code "+strconv.Itoa(code)+" and logs reason", func() {
			session := healthCheck()
			Eventually(session).Should(gexec.Exit(code))
			Expect(session.Out).To(gbytes.Say(reason))
		})
	}

	BeforeEach(func() {
		args = nil

		ip := getNonLoopbackIP()
		server = ghttp.NewUnstartedServer()
		listener, err := net.Listen("tcp", ip+":0")
		Expect(err).NotTo(HaveOccurred())

		server.HTTPTestServer.Listener = listener
		serverAddr = listener.Addr().String()
		server.Start()

		_, port, err = net.SplitHostPort(serverAddr)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("fails when parsing flags", func() {
		It("exits with code 2", func() {
			session, _ := gexec.Start(exec.Command(healthCheck, "-invalid_flag"), GinkgoWriter, GinkgoWriter)
			Eventually(session).Should(gexec.Exit(2))
		})
	})

	portHealthCheck := func() *gexec.Session {
		args = append([]string{"-port", port, "-timeout", "100ms"}, args...)
		session, err := gexec.Start(exec.Command(healthCheck, args...), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		return session
	}

	httpHealthCheck := func() *gexec.Session {
		args = append([]string{"-uri", "/api/_ping", "-port", port, "-timeout", "100ms"}, args...)
		session, err := gexec.Start(exec.Command(healthCheck, args...), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		return session
	}

	Describe("in readiness mode", func() {
		var (
			session    *gexec.Session
			statusCode int64 = http.StatusInternalServerError
		)

		BeforeEach(func() {
			server.RouteToHandler("GET", "/api/_ping", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				statusCode := atomic.LoadInt64(&statusCode)
				resp.WriteHeader(int(statusCode))
			}))

			args = []string{"-readiness-interval=1s"}
		})

		AfterEach(func() {
			session.Kill()
		})

		It("does not exit until the http server is started", func() {
			session = httpHealthCheck()
			Consistently(session).ShouldNot(gexec.Exit())
			atomic.StoreInt64(&statusCode, http.StatusOK)
			Eventually(session, 2*time.Second).Should(gexec.Exit(0))
		})

		It("runs a healthcheck every readiness-interval", func() {
			session = httpHealthCheck()
			start := time.Now()
			Eventually(server.ReceivedRequests, 3*time.Second).Should(HaveLen(2))
			end := time.Now()
			Expect(end.Sub(start)).To(BeNumerically("~", 1*time.Second, 100*time.Millisecond))
		})
	})

	Describe("in liveness mode", func() {
		var (
			session    *gexec.Session
			statusCode int64 = http.StatusOK
		)

		BeforeEach(func() {
			server.RouteToHandler("GET", "/api/_ping", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				statusCode := atomic.LoadInt64(&statusCode)
				resp.WriteHeader(int(statusCode))
			}))

			args = []string{"-liveness-interval=1s"}
		})

		AfterEach(func() {
			session.Kill()
		})

		It("does not exit until the http server is down", func() {
			session = httpHealthCheck()
			Consistently(session).ShouldNot(gexec.Exit())
			atomic.StoreInt64(&statusCode, http.StatusInternalServerError)
			Eventually(session, 2*time.Second).Should(gexec.Exit(6))
			Expect(session.Out).To(gbytes.Say("failure to get valid HTTP status code: 500"))
		})

		It("runs a healthcheck every liveness-interval", func() {
			session = httpHealthCheck()
			start := time.Now()
			Eventually(server.ReceivedRequests, 3*time.Second).Should(HaveLen(2))
			end := time.Now()
			Expect(end.Sub(start)).To(BeNumerically("~", 1*time.Second, 100*time.Millisecond))
		})
	})

	Describe("port healthcheck", func() {
		Context("when the address is listening", func() {
			itExitsWithCode(portHealthCheck, 0, "healthcheck passed")
		})

		Context("when the address is not listening", func() {
			BeforeEach(func() {
				port = "-1"
			})

			itExitsWithCode(portHealthCheck, 4, "failure to make TCP connection")
		})
	})

	Describe("http healthcheck", func() {
		Context("when the healthcheck is properly invoked", func() {
			BeforeEach(func() {
				server.RouteToHandler("GET", "/api/_ping", ghttp.VerifyRequest("GET", "/api/_ping"))
			})

			Context("when the address is listening", func() {
				itExitsWithCode(httpHealthCheck, 0, "healthcheck passed")
			})

			Context("when the address returns error http code", func() {
				BeforeEach(func() {
					server.RouteToHandler("GET", "/api/_ping", ghttp.RespondWith(500, ""))
				})

				itExitsWithCode(httpHealthCheck, 6, "failure to get valid HTTP status code: 500")
			})
		})
	})
})

func getNonLoopbackIP() string {
	interfaces, err := net.Interfaces()
	Expect(err).NotTo(HaveOccurred())
	for _, intf := range interfaces {
		addrs, err := intf.Addrs()
		if err != nil {
			continue
		}

		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	Fail("no non-loopback address found")
	panic("non-reachable")
}
