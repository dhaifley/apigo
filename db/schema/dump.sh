#!/bin/sh

docker exec -it tests-db-1 pg_dump -h localhost -U postgres -p 5432 -s api-db > db.sql
