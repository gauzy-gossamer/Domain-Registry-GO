import sys
import os
import time
import logging
import subprocess
import argparse

import psycopg2

from config import read_config, get_pg_conn

def get_ipaddr_class(ipaddr):
    return "A" if ipaddr.find(':') == -1 else "AAAA"

# generate DNS zone records
class ZoneGenerator():
    def __init__(self, iterator, output_fd=sys.stdout, header=None):
        self.iterator = iterator
        self.fd = output_fd
        self.headerfile = header

    def write_header(self):
        if self.headerfile is not None:
            with open(self.headerfile, 'r') as f:
                self.fd.write(f.read())
                self.fd.write('; ---\n\n')

    def generate_soa(self):
        soa_record = self.iterator.get_soa_record()
        soa_record['hostmaster'] = soa_record['hostmaster'].replace('@', '.')
        ns = soa_record['nameservers'][0]['nsname']
        self.fd.write(f"$TTL {soa_record['ttl']} ;default TTL for all records in zone\n")
        self.write_header()
        self.fd.write(f"{soa_record['zonename']}.\t\tIN\tSOA\t{ns}.\t{soa_record['hostmaster']}. (")
        self.fd.write(f"{soa_record['serial']} {soa_record['refresh']} {soa_record['update_retr']} ")
        self.fd.write(f"{soa_record['expiry']} {soa_record['min']})\n")

        # list of nameservers for the zone
        for ns in soa_record['nameservers']:
            self.fd.write(f"\t\tIN\tNS\t{ns['nsname']}.\n")
        # addresses of nameservers (only if there are any)
        for ns in soa_record['nameservers']:
            for addr in ns['addrs']:
                self.fd.write("{}.\tIN\t{}\t{}\n".format(ns['nsname'], get_ipaddr_class(addr), addr))

        self.fd.write(";\n\n")

    def generate_records(self):
        self.iterator.start_gen_domains()
        for (domain, hosts, dsrecords) in self.iterator.get_next_domain():
            for ns in hosts:
                self.fd.write("{}.\tIN\tNS\t{}".format(domain, ns))
                # if the nameserver's fqdn is already terminated by a dot
                # we don't add another one - ugly check which is necessary
                # becauseof error in CR (may be removed in future)
                if ns[-1] != '.':
                    self.fd.write(".\n")
                else:
                    self.fd.write("\n")

                # distinguish ipv4 and ipv6 address
                for addr in hosts[ns]:
                    self.fd.write("%s.\tIN\t%s\t%s\n" %
                            (ns, get_ipaddr_class(addr), addr))

            for ds in dsrecords:
                #if ds['maxSigLife'] > 0:
                #    ttl = ds['maxSigLife']
                #else:
                ttl = ""
                self.fd.write(f"{domain}. {ttl}\tIN\tDS\t{ds['keytag']} {ds['alg']} ")
                self.fd.write(f"{ds['digesttype']} {ds['digest']}\n")

# prepare query and iterate through records
class ZoneDB():
    def __init__(self, db, zoneid):
        self.db = db
        self.zoneid = zoneid

    def get_soa_record(self):
        cursor = self.db.cursor()
        cursor.execute("SELECT z.fqdn, zs.ttl, zs.hostmaster, zs.serial, zs.refresh,"
                "zs.update_retr, zs.expiry, zs.minimum, zs.ns_fqdn "
                "FROM zone z, zone_soa zs WHERE zs.zone = z.id AND z.id = %s",
                (self.zoneid,))
        if cursor.rowcount == 0:
            cursor.close()
            raise Exception("unknown zone")
        (zonename, ttl, hostmaster, serial, refresh, update_retr, expiry, minimum,
                ns_fqdn) = cursor.fetchone()

        if serial is None:
            serial = int(time.time())

        soa_record = {
            'zonename':zonename,
            'ttl':ttl, 'hostmaster':hostmaster,
            'serial':serial, 'refresh':refresh,
            'update_retr':update_retr, 'expiry':expiry,
            'min':minimum, 'ns':ns_fqdn,
            'nameservers':[]
        }

        # create a list of nameservers for the zone
        cursor.execute("SELECT fqdn, addrs FROM zone_ns WHERE zone = %s", (self.zoneid,))
        for (nsname, ipaddr) in cursor:
            # TODO ensure subordinate hosts have ip addresses
            soa_record['nameservers'].append({'nsname':nsname, 'addrs':ipaddr})
        cursor.close()

        return soa_record

    def start_gen_domains(self):
        zone_query = '''SELECT d.id, oreg.name, h.fqdn, a.ipaddr
            FROM object_registry oreg 
               INNER JOIN domain_host_map dh on oreg.id=dh.domainid
               JOIN host h on dh.hostid=h.hostid LEFT JOIN host_ipaddr_map a ON (h.hostid = a.hostid)
               JOIN domain d on oreg.id=d.id LEFT JOIN object_state_now osn ON (d.id = osn.object_id) 
            WHERE (NOT (15 = ANY (osn.states)) OR osn.states IS NULL) and d.zone = %s
            ORDER BY oreg.id;'''
        self.cursor = self.db.cursor()
        self.cursor.execute(zone_query, (self.zoneid,))

        # there are usually far fewer dsrecords than ns records, 
        # so we pull them with a separate query and merge with main records
        dnssec_query = '''SELECT d.id, ds.keytag, ds.alg, ds.digesttype, ds.digest
            FROM domain d JOIN dsrecord ds on ds.keysetid=d.keyset
            WHERE d.zone = %s ORDER BY d.id'''

        self.cursor2 = self.db.cursor()
        self.cursor2.execute(dnssec_query, (self.zoneid,))
        self.dsrecords = []
        # dsrecords generator
        self.iter_dnssec = self._iter_dnssec_cursor()

    # get dsrecords for next domain
    def _iter_dnssec_cursor(self):
        dsrecords = []
        cur_domid = 0
 
        while True:
            row = self.cursor2.fetchone()
            if row is None:
                yield dsrecords
                dsrecords = []
                continue
            vals = dict(zip([column[0] for column in self.cursor2.description], row))
            if cur_domid != 0 and vals['id'] != cur_domid:
                yield dsrecords
                dsrecords = []
            dsrecords.append(vals)
            cur_domid = vals['id']

    # join domains with dsrecords using merge algorithm
    def get_next_dnssec(self, domid):
        if len(self.dsrecords) > 0:
            if self.dsrecords[0]['id'] == domid:
                return self.dsrecords
            if self.dsrecords[0]['id'] > domid:
                return []
            self.dsrecords = []

        while True:
            self.dsrecords = next(self.iter_dnssec)
            if len(self.dsrecords) == 0 or self.dsrecords[0]['id'] >= domid:
                break
        if len(self.dsrecords) > 0 and self.dsrecords[0]['id'] == domid:
            return self.dsrecords

        return []
        
    def get_next_domain(self):
        cur_domain = None
        cur_domid = 0
        hosts = {}
        dsrecords = []
        for (domid, domain, host, ipaddr) in self.cursor:
            if cur_domain != domain:
                dsrecords = self.get_next_dnssec(cur_domid)
                yield (cur_domain, hosts, dsrecords)
                hosts = {}

            if host not in hosts:
                hosts[host] = []
            if ipaddr is not None:
                hosts[host].append(ipaddr)

            cur_domain = domain
            cur_domid = domid
        dsrecords = self.get_next_dnssec(cur_domid)
        yield (cur_domain, hosts, dsrecords)

def iter_zones(conn, zone=None):
    c = conn.cursor()
    c.execute('''SELECT id, fqdn FROM zone''')

    for (zoneid, fqdn) in c:
        if zone is not None and fqdn != zone:
            continue
        yield (zoneid, fqdn)

def run_checkzone(zone_filename, zone):
    process = subprocess.Popen(
               f"/usr/sbin/named-checkzone {zone} {zone_filename}", shell=True, stdout=subprocess.PIPE)

    (output, stderr) = process.communicate()
    logging.info(output.decode())
    if process.returncode != 0:
        raise Exception('checkzone failed')

def main():
    parser = argparse.ArgumentParser()

    parser.add_argument('--zone',  type=str, default=None, help="only generate zone file for this zone")
    parser.add_argument('--target-dir', dest='target_dir', type=str, default=None, help="move generated zones to this directory")
    parser.add_argument('--header',  type=str, default=None, help="prepend contents of this file to the zone")
    parser.add_argument('--run-named-checkzone', dest='run_checkzone',action='store_true', default=False, help="run named-checkzone on generated zone file")

    args = parser.parse_args()

    config = read_config()

    conn = get_pg_conn(config)

    for (zoneid, fqdn) in iter_zones(conn, args.zone):
        try:
            zone_iterator = ZoneDB(conn, zoneid)

            zone_filename = '{}.db'.format(fqdn)

            with open(zone_filename, 'w') as f:
                generator = ZoneGenerator(zone_iterator, output_fd=f, header=args.header)
                generator.generate_soa()
                generator.generate_records()

            if args.run_checkzone:
                run_checkzone(zone_filename, fqdn)

            if args.target_dir is not None:
                os.system(f"mv {zone_filename} {args.target_dir}")
        except Exception as exc:
            logging.error('zone generation failed {}\n'.format(exc))

if __name__ == '__main__':
    main()
