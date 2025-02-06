#!/bin/bash

echo "Starting test services..."

docker compose -f tests/docker-compose.yml up -d --force-recreate

echo "Setting account repository..."

wget --no-verbose \
--output-document=- \
--header 'Authorization: Bearer '"$USER_AUTH_TOKEN" \
--header 'Content-Type: application/json' \
--post-data '{"repo": "'"$REPO"'"}' \
http://localhost:8080/v1/api/account/repo

echo "Test services started, ready to run tests."
