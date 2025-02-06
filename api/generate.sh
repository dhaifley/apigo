#!/bin/sh

docker run --rm -v $PWD/..:/spec redocly/cli bundle api/index.yaml > openapi.yaml
