#! /bin/sh

registrar=$1

if [ -z $registrar ] ; then
    echo "usage $0 registrar"
    exit 1;
fi

days=2000
cert_path=client/$registrar
# passwd: abcd
ca_path=test-ca

openssl genrsa -out $cert_path.key 4096

openssl req -new -key $cert_path.key -out $cert_path.csr

openssl x509 -req -in $cert_path.csr -CA $ca_path.pem -CAkey $ca_path.key -CAcreateserial -out $cert_path.crt -days $days -sha256

fingerprint=$(openssl x509 -noout -fingerprint -md5 -inform pem -in $cert_path.crt)

echo "$registrar $fingerprint"
