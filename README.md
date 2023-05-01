# Domain-Registry-GO
Domain registry implemented in Go. The registry is functional but lacks some features (like DNSSEC, admin interfaces), and still needs work.

The core server could be set up either as a standalone process or behind nginx proxy (see config/nginx.conf).

The database schema is based on FRED registry (https://github.com/CZ-NIC/fred-db) with some changes - nssets are changed to hosts, there are tables to track sessions and transfer requests. In addition, there are changes to contact fields, calculation of object states and domain lifetimes.
