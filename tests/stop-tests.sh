#!/bin/sh

echo "Stopping test services..."

docker compose -f tests/docker-compose.yml down

echo "All test services stopped."
