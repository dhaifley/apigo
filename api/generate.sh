#!/bin/sh

docker run --rm -v $PWD:/spec redocly/cli bundle index.yaml > ../internal/server/static/openapi.yaml
docker run --rm -v $PWD:/spec redocly/cli bundle index.yaml --ext=json > ../internal/server/static/openapi.json
