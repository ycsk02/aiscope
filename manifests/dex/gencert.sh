#!/bin/bash

mkdir -p ssl

cat << EOF > ssl/req.cnf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name

[req_distinguished_name]

[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = dex.aiscope.io
EOF

openssl genrsa -out ssl/ca-key.pem 2048
openssl req -x509 -new -nodes -key ssl/ca-key.pem -days 10 -out ssl/ca.pem -subj "/CN=aiscope"

openssl genrsa -out ssl/key.pem 2048
openssl req -new -key ssl/key.pem -out ssl/csr.pem -subj "/CN=aiscope" -config ssl/req.cnf
openssl x509 -req -in ssl/csr.pem -CA ssl/ca.pem -CAkey ssl/ca-key.pem -CAcreateserial -out ssl/cert.pem -days 10 -extensions v3_req -extfile ssl/req.cnf

kubectl create ns dex
kubectl -n dex create secret tls dex.aiscope.io.tls \
  --cert=ssl/cert.pem \
  --key=ssl/key.pem
kubectl -n dex create secret generic dex.aiscope.io.ca --from-file=ssl/ca.pem

#CERT_NAME=dex.aiscope.io.tls
#HOST=dex.aiscope.io
#CERT_FILE=tls.cert
#KEY_FILE=tls.key
#openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ${KEY_FILE} -out ${CERT_FILE} -subj "/CN=${HOST}/O=${HOST}"
#kubectl create ns dex
#kubectl -n dex create secret tls ${CERT_NAME} --key ${KEY_FILE} --cert ${CERT_FILE}

#http://dex.aiscope.io:32000/auth?client_id=aiscope&redirect_uri=http://api.aiscope.io:9090/oauth/callback/dex&response_type=code&scope=openid+email+groups+profile+offline_access
