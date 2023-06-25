# Domain-Registry-GO
Domain registry implemented in Go. The registry is functional but lacks some features (like DNSSEC, admin interfaces), and still needs work.

The core server could be set up either as a standalone process or behind nginx proxy (see config/nginx.conf).

The database schema is based on FRED registry (https://github.com/CZ-NIC/fred-db) with some changes - nssets are changed to hosts, there are tables to track sessions and transfer requests. In addition, there are changes to contact fields, calculation of object states and domain lifetimes.

# Installation guide
```sh
# create db schema (sql/init.sql) and copy configuration files
./prepare_docker.sh
# run modules in docker containers
docker-compose up -d
```

# Configuration

## Set up new zone

### Add zone
```python
admin_cli.py --add-zone --zone example.com
```

### Set up SOA record
```python
admin_cli.py --set-zone-soa --zone example.com --hostmaster admin.example.com --ns-fqdn ns1.example.com
```

### Set up zone NS servers
```python
admin_cli.py --add-zone-ns --zone example.com --fqdn ns1.example.com --addrs "127.0.0.1,127.0.0.2"
admin_cli.py --add-zone-ns --zone example.com --fqdn ns2.example.com --addrs "127.0.0.1,127.0.0.2"
```

### Set operation prices
```python
admin_cli.py --zone-set-price --zone example.com --operation "CreateDomain" --price 1
admin_cli.py --zone-set-price --zone example.com --operation "RenewDomain" --price 1
admin_cli.py --zone-set-price --zone example.com --operation "TransferDomain" --price 0
```

## Set up new registrar

### Create registrar
```python
admin_cli.py --create-registrar --handle TEST-REG --name "Test Registrar" 
```
### Generate certificate
```sh
./cert/gen_registrar_cert.sh TEST-REG
```
### Set registrar ACL
```python
admin_cli.py --set-registrar-acl --handle TEST-REG --cert CERT_FINGERPRINT --password password
```

### Give access to use zones
```python
admin_cli.py --registrar-add-zone --handle TEST-REG --zone example.com
```
