package main

import (
	"flag"
	"fmt"
	"net"
	"os"
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

var readinessTimeout = flag.Duration(
	"readiness-timeout",
	60*time.Second,
	"Only relevant if healthcheck is running in readiness mode. When the timeout is set to a non-zero value, the healthcheck will return non-zero with any errors if this timeout is hit without the healthcheck passing",
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
		fmt.Fprintf(os.Stderr, "Failed to get interfaces: %s\n", err)
		os.Exit(1)
		return
	}

	h := newHealthCheck(*network, *uri, *port, *timeout)

	var timeoutTimerCh <-chan time.Time
	if duration := *readinessTimeout; duration > 0 {
		timeoutTimerCh = time.NewTimer(duration).C
	}

	if readinessInterval != nil && *readinessInterval > 0 {
		ticker := time.NewTicker(*readinessInterval)
		defer ticker.Stop()
		errCh := make(chan error)

		for attempt := 0; ; attempt++ {
			fmt.Fprintf(os.Stdout, "Readiness check attempt #%d\n", attempt)
			go func() {
				err = h.CheckInterfaces(interfaces)
				errCh <- err
			}()

			select {
			case err = <-errCh:
				if err == nil {
					fmt.Fprintf(os.Stdout, "Readiness check successful\n")
					os.Exit(0)
				}
				fmt.Fprintf(os.Stderr, "Readiness check unsuccessful: %s\n", err)
			case <-timeoutTimerCh:
				fmt.Fprintf(os.Stderr, "Timed out after %s waiting for readiness check to succeed\n", *readinessTimeout)
				failHealthCheck(err)
			}

			select {
			case <-ticker.C:
				fmt.Fprintf(os.Stderr, "Readiness Interval (%s) elapsed, checking again\n", *readinessInterval)
			case <-timeoutTimerCh:
				fmt.Fprintf(os.Stderr, "Timed out after %s waiting for readiness check to succeed\n", *readinessTimeout)
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
		fmt.Fprintf(os.Stderr, "%s\n", err.Message)
		os.Exit(err.Code)
	}

	fmt.Fprintf(os.Stderr, "Unknown error encountered in healthcheck: %s\n", err.Error())
	os.Exit(127)
}
