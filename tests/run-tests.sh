#!/bin/sh

docker run \
--network tests_default \
--volume $PWD/tests/postman:/etc/newman \
--tty postman/newman run apid.postman_collection.json \
--env-var "scheme=http" \
--env-var "host=apid:8080" \
--env-var "baseURL={{scheme}}://{{host}}/v1/api" \
--env-var "userAuthToken=$USER_AUTH_TOKEN"
