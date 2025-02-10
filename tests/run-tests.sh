#!/bin/sh

docker run \
--network tests_default \
--volume $PWD/tests/postman:/etc/newman \
--tty postman/newman run apigo.postman_collection.json \
--env-var "scheme=http" \
--env-var "host=apigo:8080" \
--env-var "baseURL={{scheme}}://{{host}}/api/v1" \
--env-var "userAuthToken=$USER_AUTH_TOKEN"
