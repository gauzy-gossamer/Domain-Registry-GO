import sys 
import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

from epp_client import create_epp_client
from dsrecord import DSRecord

settings = {
    'registrar': "TEST-REG",
    'passwd':"password",
    'test_contact':"test-person",
    'test_host':'ns1.syst.test.com',
    'zone':'.ex.com',
    'dnskey_algo':12,
    'dnskey_pubkey':'LMgXRHzSbIJGn6i16K+sDjaDf/k1o9DbxScOgEYqYS/rlh2Mf+BRAY3QHPbwoPh2fkDKBroFSRGR7ZYcx+YIQw==',
    'digest_algo':1, # SHA1
}

def run_main():
    cert = "client.crt"
    key = "client.key"

    client = create_epp_client(settings['registrar'], settings['passwd'], lang='en',
                    cert=cert, key=key, server='localhost', port=8090)

    domain = '33400412' + settings['zone']

    secdns = DSRecord().generate_dsrecord(domain, settings['dnskey_pubkey'], algorithm=settings['dnskey_algo'], digest_algo=settings['digest_algo'])

    try:
        client.create_domain(domain, registrant=settings['test_contact'], secdns=[secdns])

        client.info_domain(domain)

        client.update_domain(domain, rem_secdns='all')

        client.delete_domain(domain)
    except Exception as exc:
        print(exc)

    client.logout()

if __name__ == '__main__':
    run_main()
