#!/bin/sh

echo "Building test services..."

docker compose -f tests/docker-compose.yml build 

echo "Build complete, ready to start tests."
