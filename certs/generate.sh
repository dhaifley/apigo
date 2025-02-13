#!/usr/bin/sh

# Certificate Authority (CA)
openssl req -x509 -newkey rsa:4096 -days 365 -nodes \
    -keyout certs/ca.key -out certs/ca.crt \
    -subj "/C=US/O=apigo/CN=apigo CA"

# Server certificate request
openssl req -newkey rsa:4096 -nodes \
    -keyout certs/tls.key -out certs/tls.csr \
    -subj "/C=US/O=apigo/CN=localhost"

# Server certificate signed with the CA
openssl x509 -req -in certs/tls.csr -days 365 \
    -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial \
    -out certs/tls.crt
