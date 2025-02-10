#!/bin/sh

echo "Stopping test services..."

docker compose -f tests/docker-compose.yml down --remove-orphans --volumes

echo "All test services stopped."
