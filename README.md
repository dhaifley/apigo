# apid
An application programming interface service template

A set of packages that may be useful as a template project when creating
services which expose a REST API using Go.

## Requirements

* [go](https://go.dev/dl/)
* [docker](https://docs.docker.com/get-docker/)

## Building and Testing

To setup a local test environment, from the project root directory, run:

```sh
$ tests/build-tests.sh
```

This will build container images for a local test environment.

Then, you must set the `USER_AUTH_TOKEN` environment variable to be a valid API
authentication token. A sample token, created with the sample test certificates
is provided in the tests/test.env file

```sh
$ set -a

$ source tests/test.env
```

To start the local test environment containers, run:

```sh
$ tests/start-tests.sh
```

To see the log output for the various test containers, run:

```sh
$ docker compose logs
```

To execute the integration tests, run:

```sh
$ tests/run-tests.sh
```

To shutdown and cleanup the test environment, run:

```sh
$ tests/stop-tests.sh
```

While the local test environment is running, the test API service can be
accessed at: `http://localhost:8080/v1/api`.

## API Documentation

* [OpenAPI spec documentation](api/openapi.yaml)
