#!/bin/bash

python=python3

# copy registry conf
$python scripts/build_template.py config/xsd_schemas registry
$python scripts/build_template.py config/server.conf registry/server.conf.example
cp cert/server/server.key registry/test-key.pem
cp cert/server/server.crt registry/test-cert.pem

# copy whois conf
cp config/server.conf whois/server.conf.example

./build_db_schema.sh --with-test > sql/init.sql
