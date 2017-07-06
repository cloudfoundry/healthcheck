package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"code.cloudfoundry.org/healthcheck"
)

var network = flag.String(
	"network",
	"tcp",
	"network type to dial with (e.g. unix, tcp)",
)

var uri = flag.String(
	"uri",
	"",
	"uri to healthcheck",
)

var port = flag.String(
	"port",
	"8080",
	"port to healthcheck",
)

var timeout = flag.Duration(
	"timeout",
	1*time.Second,
	"dial timeout",
)

var readinessInterval = flag.Duration(
	"readiness-interval",
	0,
	"if set, starts the healthcheck in readiness mode, i.e. do not exit until the healthcheck passes. runs checks every readiness-interval",
)

var livenessInterval = flag.Duration(
	"liveness-interval",
	0,
	"if set, starts the healthcheck in liveness mode, i.e. do not exit until the healthcheck fail. runs checks every liveness-interval",
)

func main() {
	flag.Parse()

	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(fmt.Sprintf("failure to get interfaces: %s", err))
		os.Exit(1)
		return
	}

	h := newHealthCheck(*network, *uri, *port, *timeout)

	if readinessInterval != nil && *readinessInterval > 0 {
		ticker := time.NewTicker(*readinessInterval)
		defer ticker.Stop()
		sigCh := make(chan os.Signal)
		errCh := make(chan error)
		signal.Notify(sigCh, syscall.SIGTERM)

		for {
			go func() {
				errCh <- h.CheckInterfaces(interfaces)
			}()

			select {
			case err = <-errCh:
				if err == nil {
					os.Exit(0)
				}
			case <-sigCh:
				failHealthCheck(err)
			}

			select {
			case <-ticker.C:
			case <-sigCh:
				failHealthCheck(err)
			}
		}
	}

	if livenessInterval != nil && *livenessInterval > 0 {
		for {
			err = h.CheckInterfaces(interfaces)
			if err != nil {
				failHealthCheck(err)
			}
			time.Sleep(*livenessInterval)
		}
	}

	err = h.CheckInterfaces(interfaces)
	if err == nil {
		os.Exit(0)
	}

	failHealthCheck(err)
}

func failHealthCheck(err error) {
	if err, ok := err.(healthcheck.HealthCheckError); ok {
		fmt.Print(err.Message)
		os.Exit(err.Code)
	}

	fmt.Print("healthcheck failed(unknown error)" + err.Error())
	os.Exit(127)
}
