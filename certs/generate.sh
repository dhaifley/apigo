rm *.crt

rm *.key

openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca.key -out ca.crt

echo "CA certificate"

openssl x509 -in ca.crt -noout -text

openssl req -newkey rsa:4096 -nodes -keyout tls.key -out tls.csr

openssl x509 -req -in tls.csr -days 365 -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt

echo "Signed certificate"

openssl x509 -in tls.crt -noout -text
