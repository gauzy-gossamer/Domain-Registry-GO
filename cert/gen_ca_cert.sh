#!/bin/bash

BOLD=$(tput bold)
CLEAR=$(tput sgr0)

iterate=(server/ client/)
for dir in "${iterate[@]}"; do
  [[ ! -d "$dir" ]] && mkdir -p "$dir" \
  && echo -e "${BOLD}directory '$dir' was created ${CLEAR}"
done

days=2000
ca_path=test-ca
server_path=server/server

echo -e "${BOLD}Generating RSA AES-256 Private Key for Root Certificate Authority${CLEAR}"
openssl genrsa -aes256 -out $ca_path.key 4096

echo -e "${BOLD}Generating Certificate for Root Certificate Authority${CLEAR}"
openssl req -x509 -new -nodes -key $ca_path.key -sha256 -days $days -out $ca_path.pem

echo -e "${BOLD}Generating RSA Private Key for Server Certificate${CLEAR}"
openssl genrsa -out $server_path.key 4096

echo -e "${BOLD}Generating Certificate Signing Request for Server Certificate${CLEAR}"
openssl req -new -key $server_path.key -out $server_path.csr

echo -e "${BOLD}Generating Certificate for Server Certificate${CLEAR}"
openssl x509 -req -in $server_path.csr -CA $ca_path.pem -CAkey $ca_path.key -CAcreateserial -out $ca_path.crt -days $days -sha256

echo "Done!"
