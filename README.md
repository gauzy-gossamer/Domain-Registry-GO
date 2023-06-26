# Domain-Registry-GO
Domain registry implemented in Go. The registry is functional but lacks some features (like DNSSEC, admin interfaces), and still needs work.

The core server could be set up either as a standalone process or behind nginx proxy (see config/nginx.conf).

The database schema is based on FRED registry (https://github.com/CZ-NIC/fred-db) with some changes - nssets are changed to hosts, there are tables to track sessions and transfer requests. In addition, there are changes to contact fields, calculation of object states and domain lifetimes.

# Installation guide
```sh
# create db schema (sql/init.sql) and copy configuration files
./prepare_docker.sh
```

By default database schema is generated with two registrars (SYSTEM-REG, TEST-REG) and a test zone (ex.com). Use "./build_db_schema.sh > sql/init.sql" to generate schema with predefined registrars and zones.

```sh
# run modules in docker containers
$ docker-compose up -d

$ docker ps
CONTAINER ID   IMAGE                   COMMAND                  CREATED              STATUS                        PORTS                                                                                      NAMES
8f69c96c517c   registry_regcore        "/usr/src/registry/r…"   About a minute ago   Up About a minute             0.0.0.0:8090->8090/tcp, :::8090->8090/tcp, 0.0.0.0:51015->51015/tcp, :::51015->51015/tcp   regcore
ee2b4a51c33e   registry_regadmin       "uvicorn main:app --…"   About a minute ago   Up About a minute (healthy)   0.0.0.0:8088->8088/tcp, :::8088->8088/tcp                                                  regadmin
6c6cd0386e44   registry_regwhois       "/usr/src/whois/whois"   About a minute ago   Up About a minute (healthy)   0.0.0.0:8043->8043/tcp, :::8043->8043/tcp                                                  regwhois
eb1b9f569788   postgres:latest         "docker-entrypoint.s…"   About a minute ago   Up About a minute (healthy)   0.0.0.0:7432->5432/tcp, :::7432->5432/tcp                                                  registry_regdb_1
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

Use script cert/gen_registrar_cert.sh to generate registrar certificates.
Note that for production environment you need to generate CA certificates first with gen_ca_cert.sh command.

```sh
# this will create files client/TEST-REG.crt client/TEST-REG.csr client/TEST-REG.key
$ ./gen_registrar_cert.sh TEST-REG
...
TEST-REG md5 Fingerprint=05:3D:AE:11:5C:CC:A6:C1:08:61:24:E1:EA:12:67:50
```

### Set registrar ACL
Set up access parameters. Use md5 fingerprint from the generated certificate as CERT_FINGERPRINT.
```python
admin_cli.py --set-registrar-acl --handle TEST-REG --cert CERT_FINGERPRINT --password password
```

### Give access to use zones
```python
admin_cli.py --registrar-add-zone --handle TEST-REG --zone example.com
```

## Regular tasks

Set up these jobs with crontab.

### Zone generation

This command will create a file ex.com.db with a generated DNS zone:

```python
cd scripts && python3 generate_zone.py --zone ex.com
```

### Deletion of expired domains

Script connects to registry core using gRPC service (by default 127.0.0.1:51015).

```sh
cd scripts && python3 remove_expired_domains.py
```

## Test EPP Client

Install EPPy module:
```sh
git clone https://github.com/gauzy-gossamer/eppy ./
cd eppy && python3 setup.py install
```

Now you can use python EPP client to test how registry works:

```python
# create client
>>> import scripts.epp_client as epp_client
>>> client = epp_client.create_epp_client('TEST-REG', 'password', cert='cert/client/test-client.crt', key='cert/client/test-client.key', server='localhost', port=8090)
<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
  <command>
    <login>
      <clID>TEST-REG</clID>
      <pw>password</pw>
      <options>
        <version>1.0</version>
        <lang>en</lang>
      </options>
      <svcs>
        <objURI>urn:ietf:params:xml:ns:domain-1.0</objURI>
        <objURI>urn:ietf:params:xml:ns:host-1.0</objURI>
        <objURI>urn:ietf:params:xml:ns:contact-1.0</objURI>
      </svcs>
    </login>
  </command>
</epp>

<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
  <response>
    <result code="1000">
      <msg>Command completed successfully</msg>
    </result>
    <trID>
      <clTRID>2nteQ20vWyF6</clTRID>
      <svTRID>SV-XpqwEpBUTH</svTRID>
    </trID>
  </response>
</epp>

# create test contact
>>> contact_info = {
...     'intPostalInfo':{'name':'John Snow'},
...     'locPostalInfo':{'name':'John Snow','address':'Winterfell'},
...     'birthday':'2000-01-01',
...     'passport':['passport info'], 
...     'voice':['+1 900 0000'],
...     'fax': ['+1 900 0001'],
...     'email':['john.snow@gmail.com'],
... }
>>> client.create_contact('test-person', contact_type='person', fields=contact_info)
<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
  <command>
    <create>
      <contact:create xmlns:contact="http://www.ripn.net/epp/ripn-contact-1.0" xmlns="http://www.ripn.net/epp/ripn-contact-1.0">
        <id>test-person</id>
        <person>
          <intPostalInfo>
            <name>John Snow</name>
          </intPostalInfo>
          <locPostalInfo>
            <name>John Snow</name>
            <address>Winterfell</address>
          </locPostalInfo>
          <birthday>2000-01-01</birthday>
          <passport>passport info</passport>
          <voice>+1 900 0000</voice>
          <fax>+1 900 0001</fax>
          <email>john.snow@gmail.com</email>
        </person>
      </contact:create>
    </create>
  </command>
</epp>

<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
  <response>
    <result code="1000">
      <msg>Command completed successfully</msg>
    </result>
    <resData>
      <contact:creData xmlns:contact="http://www.ripn.net/epp/ripn-contact-1.0" xmlns="http://www.ripn.net/epp/ripn-contact-1.0">
        <id>test-person</id>
        <crDate>2023-06-26T20:56:37Z</crDate>
      </contact:creData>
    </resData>
    <trID>
      <clTRID>yLZuHKZNFweN</clTRID>
      <svTRID>SV-pKhqHKtswC</svTRID>
    </trID>
  </response>
</epp>

# create test domain
>>> client.create_domain('domain.ex.com', registrant='test-person')
<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
  <command>
    <create>
      <domain:create xmlns:domain="http://www.ripn.net/epp/ripn-domain-1.0" xmlns="http://www.ripn.net/epp/ripn-domain-1.0">
        <name>domain.ex.com</name>
        <period unit="y">1</period>
        <registrant>test-person</registrant>
      </domain:create>
    </create>
  </command>
</epp>

<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
  <response>
    <result code="1000">
      <msg>Command completed successfully</msg>
    </result>
    <resData>
      <domain:creData xmlns:domain="http://www.ripn.net/epp/ripn-domain-1.0" xmlns="http://www.ripn.net/epp/ripn-domain-1.0">
        <name>domain.ex.com</name>
        <crDate>2023-06-26T20:58:38Z</crDate>
        <exDate>2024-06-26T20:58:38Z</exDate>
      </domain:creData>
    </resData>
    <trID>
      <clTRID>5V4Uj32OMDth</clTRID>
      <svTRID>SV-EgGCcGzMmF</svTRID>
    </trID>
  </response>
</epp>

# logout
>>> client.logout()
<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
  <command>
    <logout />
  </command>
</epp>

<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
  <response>
    <result code="1500">
      <msg>Command completed successfully; ending session</msg>
    </result>
    <trID>
      <svTRID>SV-yOGCVDZdol</svTRID>
    </trID>
  </response>
</epp>
```
