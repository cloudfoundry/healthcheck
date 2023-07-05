# healthcheck

**Note**: This repository should be imported as `code.cloudfoundry.org/healthcheck`.

Common healthcheck for buildpacks and docker.

## Reporting issues and requesting features

Please report all issues and feature requests in [cloudfoundry/diego-release](https://github.com/cloudfoundry/diego-release/issues).

## Types of Healthchecks

### Startup Healthcheck

```
# HTTP Startup Healthcheck
./healthcheck -uri=URI \
     -startup-interval=INTERVAL \
     [-port=PORT]
     [-timeout=TIMEOUT] \
     [-startup-timeout=STARTUP_TIMEOUT]

# TCP Startup Healthcheck
./healthcheck \
     -startup-interval=INTERVAL \
     [-port=PORT]
     [-timeout=TIMEOUT] \
     [-startup-timeout=STARTUP_TIMEOUT]
```

| Flag | Default | Description |
|---|---|---|
| uri | no default | URI to healthcheck. Required for HTTP healthchecks. |
| port | 8080 | Port to healthcheck.  |
| timeout | 1s  | Dial timeout when connecting to app. |
| startup-interval  | 0s | If set, starts the healthcheck in startup mode, i.e. do not exit until the healthcheck passes. Runs checks every startup-interval. Required for startup healthchecks. |
| startup-timeout  | 60s  | Only relevant if healthcheck is running in startup mode. When the timeout is set to a non-zero value, the healthcheck will return non-zero with any errors if this timeout is hit without the healthcheck passing. |

The startup healthcheck should be used when an app is starting up. It will return zero when the healthcheck gets a successful response. It will return non-zero when it does not get a successful response within the timeouts; this means that the app did not start in the timeout provided.

### Liveness Healthcheck

```
# HTTP Liveness Healthcheck
./healthcheck -uri=URI \
     -liveness-interval=INTERVAL \
     [-port=PORT]
     [-timeout=TIMEOUT]

# TCP Liveness Healthcheck
./healthcheck \
     -liveness-interval=INTERVAL \
     [-port=PORT]
     [-timeout=TIMEOUT]
```

| Flag | Default | Description |
|---|---|---|
| uri | no default | URI to healthcheck. Required for HTTP healthchecks. |
| port | 8080 | Port to healthcheck.  |
| timeout | 1s  | Dial timeout when connecting to app. |
| liveness-interval | 0s | If set, starts the healthcheck in liveness mode, i.e. the app is alive and hasn't crashed, do not exit until the healthcheck fails. runs checks every liveness-interval. Required for liveness healthchecks. |

The Liveness healthcheck should be used once the app has passed the startup healthcheck. This healthcheck will return non-zero when the healthcheck gets a failure response. This indicates that the app was running, but something has gone wrong. As long as the healthcheck keeps getting a healthy response from the app, then it will not stop running.

### Until Ready Readiness Healthcheck

```
# HTTP Until Ready Readiness Healthcheck
./healthcheck -uri=URI \
     -until-ready-interval=INTERVAL \
     [-port=PORT]
     [-timeout=TIMEOUT]

# TCP Until Ready Readiness Healthcheck
./healthcheck \
     -until-ready-interval=INTERVAL \
     [-port=PORT]
     [-timeout=TIMEOUT]
```

| Flag | Default | Description |
|---|---|---|
| uri | no default | URI to healthcheck. Required for HTTP healthchecks. |
| port | 8080 | Port to healthcheck.  |
| timeout | 1s  | Dial timeout when connecting to app. |
| until-ready-interval | 0s | If set, starts the healthcheck in until-ready mode, i.e. do not exit until the healthcheck passes and the app is ready to serve traffic. Runs checks every until-ready-interval. Required for until ready readiness healthchecks. |

The until ready readiness healthcheck will return zero when the healthcheck gets a successful response. This indicates that the app is running and ready to be routed to. As long as the healthcheck keeps getting a failure response from the app, then it will not stop running.


### Until Failure Readiness Healthcheck

```
# HTTP Until Failure Readiness Healthcheck
./healthcheck -uri=URI \
     -readiness-interval=INTERVAL \
     [-port=PORT]
     [-timeout=TIMEOUT]

# TCP Until Failure Readiness Healthcheck
./healthcheck \
     -readiness-interval=INTERVAL \
     [-port=PORT]
     [-timeout=TIMEOUT]
```

| Flag | Default | Description |
|---|---|---|
| uri | no default | URI to healthcheck. Required for HTTP healthchecks. |
| port | 8080 | Port to healthcheck.  |
| timeout | 1s  | Dial timeout when connecting to app. |
| readiness-interval | 0s | If set, starts the healthcheck in readiness mode, i.e. the app is ready to serve traffic, i.e. do not exit until the healthcheck fails because the target isn't serving traffic or another process doesn't exist. Runs checks every readiness-interval. Required for until failure readiness healthchecks. |

The until ready failure healthcheck will return non-zero when the healthcheck gets a failure response. This indicates that the app is no longer ready to be routed to. As long as the healthcheck keeps getting a success response from the app, then it will not stop running.
