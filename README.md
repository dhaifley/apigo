# apigo
An application programming interface service template.

This set of packages may be useful as a template project when creating services
which expose a REST API using Go.

## Requirements

* [go](https://go.dev/dl/)
* [docker](https://docs.docker.com/get-docker/)

## Building and Testing

To build containers for testing the service locally:

```sh
$ tests/build-tests.sh
```

Then, to start the test environment containers:

```sh
$ tests/start-tests.sh
```

To see the log output for the various test containers:

```sh
$ docker compose logs
```

To run the integration tests:

```sh
$ tests/run-tests.sh
```

Finally, to shutdown and cleanup the test environment:

```sh
$ tests/stop-tests.sh
```

## Documentation

While the local test environment is running, interactive documentation can be accessed using:
* http://localhost:8080/api/v1/docs
