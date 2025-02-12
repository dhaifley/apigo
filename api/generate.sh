#!/bin/sh

docker run --rm -v $PWD/api:/spec redocly/cli bundle index.yaml > ./internal/static/openapi.yaml
docker run --rm -v $PWD/api:/spec redocly/cli bundle index.yaml --ext=json > ./internal/static/openapi.json
