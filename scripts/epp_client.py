import copy
import time

from eppy.client import EppClientHTTP
import eppy.doc as ec
import xmltodict

class RipnEpp():
    def __init__(self, client, print_commands=True, nsmap_url='http://www.ripn.net', registrar=None, timing=False) -> None:
        self.nsmap_url = nsmap_url
        self.extra_nsmap = {
            '':'{}/epp/ripn-epp-1.0'.format(nsmap_url),
            'epp':'{}/epp/ripn-epp-1.0'.format(nsmap_url),
            'domain':'{}/epp/ripn-domain-1.0'.format(nsmap_url),
            'contact':'{}/epp/ripn-contact-1.0'.format(nsmap_url),
            'host':'{}/epp/ripn-host-1.0'.format(nsmap_url),
            'registrar':'{}/epp/ripn-registrar-1.0'.format(nsmap_url)
        }

        self.authinfo_transfer = False
        self.session_open = False
        self.timing = timing

        self.client = client
        self.registrar = registrar
        self.print_commands = print_commands

        # order of fields in xml queries, so that they pass schema validation
        contact_order = ('intPostalInfo', 'locPostalInfo', 'legalInfo', 'taxpayerNumbers', 'birthday', 'passport', 'voice', 'fax', 'email', 'authInfo', 'disclose')
        postal_order = ('name', 'org', 'address')
        person_order = {'__order':contact_order, 'intPostalInfo':{'__order':postal_order}, 'locPostalInfo':{'__order':postal_order}}
        org_order = {'__order':contact_order, 'intPostalInfo':{'__order':postal_order}, 'locPostalInfo':{'__order':postal_order}, 'legalInfo':{'__order':postal_order}}

        ec.EppCreateContactCommand._childorder['__order'] = ('id', 'person', 'organization', 'verified', 'unverified', 'authInfo')
        ec.EppCreateContactCommand._childorder['organization'] = org_order
        ec.EppCreateContactCommand._childorder['person'] = person_order

        ec.EppUpdateContactCommand._childorder['__order'] = ('id', 'add', 'rem', 'chg')
        ec.EppUpdateContactCommand._childorder['chg']['organization'] = org_order
        ec.EppUpdateContactCommand._childorder['chg']['person'] = org_order

        ec.EppCreateDomainCommand._childorder['__order'] = ('name', 'period', 'ns', 'registrant', 'description', 'authInfo')
        ec.EppUpdateDomainCommand._childorder['__order'] = ('name', 'add', 'rem', 'chg')

#        ec.EppUpdateRegistrarCommand._childorder['chg'] = {'__order':('voice', 'fax', 'www', 'whois')}

    def set_print_commands(self, print_commands: bool) -> str:
        self.print_commands = print_commands

    def set_registrar(self, registrar: str) -> None:
        self.registrar = registrar

    def set_authinfo_transfer(self, authinfo_transfer: bool) -> None:
        self.authinfo_transfer = authinfo_transfer

    def _client_send(self, cmd) -> dict:
        start_tm = time.time()
        if self.print_commands:
            print("command:")
            print(cmd)
        ret = self.client.send(cmd, extra_nsmap=self.extra_nsmap)
        if self.timing:
            print('time = {}'.format(time.time() - start_tm))
        if self.print_commands:
            print(ret)
        return xmltodict.parse(str(ret))

    def hello(self):
        cmd = ec.EppHello(extra_nsmap=self.extra_nsmap)

        return self._client_send(cmd)

    def login(self, clid: str, pw: str, lang='en', clTRID: str=None, newpw: str=None) -> dict:
        cmd = ec.EppLoginCommand(extra_nsmap=self.extra_nsmap, lang=lang)

        if clTRID is not None:
            cmd.add_clTRID(clTRID)

        cmd.clID = clid
        cmd.pw = pw

        if newpw is not None:
            cmd.newPW = newpw

        ret = self._client_send(cmd)
        self.session_open = True
        return ret

    def logout(self) -> dict:
        cmd = ec.EppLogoutCommand(extra_nsmap=self.extra_nsmap)
        return self._client_send(cmd)

    def poll(self, op='req', msg:str = None) -> dict:
        cmd = ec.EppPollCommand(op, msgID=msg, extra_nsmap=self.extra_nsmap)
        return self._client_send(cmd)

    def info_contact(self, client_handle: str, authinfo=None) -> dict:
        cmd = ec.EppInfoContactCommand(extra_nsmap=self.extra_nsmap)
        cmd.id = client_handle

        if authinfo is not None:
            cmd.authInfo = {'pw':authinfo}

        return self._client_send(cmd)

    def check_contact(self, contact: list[str]):
        cmd = ec.EppCheckContactCommand(extra_nsmap=self.extra_nsmap)
        cmd.id = contact
        return self._client_send(cmd)

    def _set_default(self, fields, field_name, def_value):
        if fields is not None and field_name in fields:
            if type(fields[field_name]) is type(def_value):
                return fields[field_name]
            else:
                raise Exception('types dont match {} {}'.format(type(fields[field_name]), type(def_value)))
        return def_value

    def create_contact(self, client_handle: str, contact_type='org', authinfo=None, fields=None) -> dict:
        cmd = ec.EppCreateContactCommand(extra_nsmap=self.extra_nsmap)
        cmd.id = client_handle

        if contact_type == 'org':
            cmd.organization = fields
        else:
            cmd.person = fields
        if authinfo is not None:
            cmd.authInfo = {'pw':authinfo}

        if fields is not None and 'verified' in fields and fields['verified']:
            cmd.verified = True

        return self._client_send(cmd)

    def update_contact(self, client_handle: str, contact_type='org', authinfo=None, add_status=None, rem_status=None, fields=None) -> dict:
        cmd = ec.EppUpdateContactCommand(extra_nsmap=self.extra_nsmap)
        cmd.id = client_handle
        person = {}
        if fields is not None:
            person = fields

        if contact_type == 'org':
            cmd.chg = {'organization':person}
        else:
            cmd.chg = {'person':person}

        if authinfo is not None:
            cmd.chg.authInfo = {'pw':authinfo}

        if add_status is not None:
            cmd.add = {'status':[{'@s':'clientUpdateProhibited'}]}
        if rem_status is not None:
            cmd.rem = {'status':[{'@s':'clientUpdateProhibited'}]}

        return self._client_send(cmd)

    def delete_contact(self, client_handle: str) -> dict:
        cmd = ec.EppDeleteContactCommand(extra_nsmap=self.extra_nsmap)
        cmd.id = client_handle
        return self._client_send(cmd)

    def _idna_domain(self, domain):
        domain = domain.lower()

        if type(domain) is str: #unicode:
            try:
                return domain.encode('idna').decode('utf-8')
            except UnicodeError:
                return domain
        else:
            return domain

    def info_domain(self, domain: str, authinfo=None, clTRID=None) -> dict:
        cmd = ec.EppInfoDomainCommand(extra_nsmap=self.extra_nsmap)
        cmd.name = self._idna_domain(domain)
        if clTRID is not None:
            cmd.add_clTRID(clTRID)

        if authinfo is not None:
            cmd.authInfo = {'pw':authinfo}

        return self._client_send(cmd)

    def check_domain(self, domains: list[str], clTRID=None) -> dict:
        cmd = ec.EppCheckDomainCommand(extra_nsmap=self.extra_nsmap)
        cmd.name = [self._idna_domain(domain) for domain in domains]
        if clTRID is not None:
            cmd.add_clTRID(clTRID)
        return self._client_send(cmd)

    def create_domain(self, domain: str, registrant: str, host=None, description=None, secdns=None, authinfo=None) -> dict:
        cmd = ec.EppCreateDomainCommand(extra_nsmap=self.extra_nsmap)
        cmd.name = self._idna_domain(domain)
        if host is not None:
            cmd.ns = {'hostObj':host}
        if description is not None:
            cmd.description = description
        cmd.registrant = registrant
        cmd.period = {'@unit':'y', '_text':'1'}
        if authinfo is not None:
            cmd.authInfo = {'pw':authinfo}
        if secdns is not None:
            cmd.add_secdns_data(secdns)

        return self._client_send(cmd)

    def renew_domain(self, domain: str, curExpDate) -> dict:
        cmd = ec.EppRenewDomainCommand(extra_nsmap=self.extra_nsmap)
        cmd.name = self._idna_domain(domain)
        cmd.curExpDate = curExpDate
        cmd.period = {'@unit':'y', '_text':'1'}

        return self._client_send(cmd)

    def update_domain(self, domain: str, registrant=None, hosts_add=None, hosts_rem=None, add_status=None, rem_status=None, description=None,
                      add_secdns=None, rem_secdns=None, authinfo=None) -> dict:
        cmd = ec.EppUpdateDomainCommand(extra_nsmap=self.extra_nsmap)
        cmd.name = self._idna_domain(domain)
        if hosts_add is not None and len(hosts_add) > 0:
            cmd.add = {'ns':{'hostObj':hosts_add}}
        if hosts_rem is not None and len(hosts_rem) > 0:
            cmd.rem = {'ns':{'hostObj':hosts_rem}}
        if registrant is not None:
            cmd.chg = {'registrant':registrant}
        if description is not None:
            cmd.chg = {'description':description}
        if add_status is not None:
            cmd.add = {'status':[{'@s':add_status}]}
        if rem_status is not None:
            cmd.rem = {'status':[{'@s':rem_status}]}

        if authinfo is not None:
            cmd.chg = {'authInfo':{'pw':authinfo}}

        secdns = {}
        if add_secdns is not None:
            if type(add_secdns) is not list:
                raise ValueError("incorrect add_secdns value")

            secdns['add'] = add_secdns
        if rem_secdns is not None:
            if type(rem_secdns) is not list:
                if rem_secdns != 'all':
                    raise ValueError("incorrect rem_secdns value")
                secdns['rem'] = [{'type':'all', 'value':'true'}]
            else:
                secdns['rem'] = rem_secdns

        if len(secdns) > 0:
            cmd.add_secdns_data(secdns)

        return self._client_send(cmd)

    def delete_domain(self, domain):
        cmd = ec.EppDeleteDomainCommand(extra_nsmap=self.extra_nsmap)
        cmd.name = self._idna_domain(domain)

        return self._client_send(cmd)

    def transfer_domain(self, domain, acquirer_id, authinfo=None, op='request'):
        extra_nsmap = copy.deepcopy(self.extra_nsmap)
        if self.authinfo_transfer:
            extra_nsmap['domain'] = '{}/epp/ripn-domain-1.1'.format(self.nsmap_url)
        cmd = ec.EppTransferDomainCommand(op, extra_nsmap=extra_nsmap)
        cmd.name = self._idna_domain(domain)
        if op == 'request' or op == 'approve' or op == 'query':
            if acquirer_id is not None:
                cmd.acID = acquirer_id

        return self._client_send(cmd)

    def info_host(self, host):
        cmd = ec.EppInfoHostCommand(extra_nsmap=self.extra_nsmap)
        cmd.name = host
        return self._client_send(cmd)

    def check_host(self, host):
        cmd = ec.EppCheckHostCommand(extra_nsmap=self.extra_nsmap)
        cmd.name = host
        return self._client_send(cmd)

    def create_host(self, host, addr=None):
        cmd = ec.EppCreateHostCommand(extra_nsmap=self.extra_nsmap)
        cmd.name = host
        if addr is not None:
            cmd.addr = addr
        return self._client_send(cmd)

    def update_host(self, host, add_addr=None, new_name=None, rem_addr=None, add_status=None, rem_status=None):
        cmd = ec.EppUpdateHostCommand(extra_nsmap=self.extra_nsmap)
        cmd.name = host
        if add_addr is not None:
            cmd.add = {'addr':add_addr}
        if rem_addr is not None:
            cmd.rem = {'addr':rem_addr}
        if new_name is not None:
            cmd.chg = {'name':new_name}

        if add_status is not None:
            cmd.add = {'status':[{'@s':'clientUpdateProhibited'}]}
        if rem_status is not None:
            cmd.rem = {'status':[{'@s':'clientUpdateProhibited'}]}
        return self._client_send(cmd)

    def delete_host(self, host):
        cmd = ec.EppDeleteHostCommand(extra_nsmap=self.extra_nsmap)
        cmd.name = host
        return self._client_send(cmd)

    def info_registrar(self, registrar):
        cmd = ec.EppInfoRegistrarCommand(extra_nsmap=self.extra_nsmap)
        cmd.id = registrar
        return self._client_send(cmd)

    def update_registrar(self, registrar: str, add_addr=None, rem_addr=None) -> dict:
        cmd = ec.EppUpdateRegistrarCommand(extra_nsmap=self.extra_nsmap)
        cmd.id = registrar
        cmd.chg = {}

        if add_addr is not None:
            cmd.add = {'addr':add_addr}

        if rem_addr is not None:
            cmd.rem = {'addr':rem_addr}

        return self._client_send(cmd)

def create_epp_client(registrar: str, passwd: str, lang: str='en', cert=None, key=None, server='127.0.0.1', port:int =700, print_commands=True) -> RipnEpp:
    client = EppClientHTTP(port=port, host=server, ssl_certfile=cert,
                     ssl_keyfile=key, ssl_cacerts=False, extra_nsmap={'':'http://www.ripn.net/epp/ripn-epp-1.0'})

    client.connect(server, port=port)

    ripn_client = RipnEpp(client=client, timing=False, print_commands=print_commands)

    ret = ripn_client.login(registrar, passwd, lang=lang)
    if ret['epp']['response']['result']['@code'] != '1000':
        raise Exception('could not login')

    return ripn_client
