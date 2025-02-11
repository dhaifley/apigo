#!/bin/sh

docker run --rm -v $PWD:/spec redocly/cli bundle index.yaml > ../internal/static/openapi.yaml
docker run --rm -v $PWD:/spec redocly/cli bundle index.yaml --ext=json > ../internal/static/openapi.json
