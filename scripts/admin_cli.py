import logging
import argparse
from datetime import date
import asyncio
import json
import aiohttp

from config import read_config

class AdminAPI():
    def __init__(self, path : str):
        self.path = path

    async def _call(self, call : str, params : dict, method : str='get') -> dict:
        path = '{}{}'.format(self.path, call)
        try:
            async with aiohttp.ClientSession(connector=aiohttp.TCPConnector(ssl=False)) as session:
                if method == 'get':
                    async with session.request(method, path, params=params) as response:
                        data = await response.json()
                else:
                    async with session.request(method, path, json=params) as response:
                        data = await response.json(content_type=None)
        except Exception as exc:
            logging.error(exc)
            return None
        return data

    async def list_zones(self, zone : str = None):
        params = {}
        if zone is not None:
            params['zone'] = zone
        return await self._call("/zones", params)

    async def add_zone(self, fqdn : str, ex_period_max : int = 12, ex_period_min : int = 12):
        return await self._call("/zones", {'fqdn':fqdn, 'ex_period_max':ex_period_max, 'ex_period_min':ex_period_min}, method='post')

    async def get_zone_soa(self, zoneid : int):
        return await self._call(f"/zones/{zoneid}/soa", {})

    async def set_zone_soa(self, zoneid: int, serial: int, ttl: int, refresh: int, update_retr: int, expiry: int, minimum: int, hostmaster: str, ns_fqdn: str):
        params = {
            'ttl': ttl,
            'serial': serial,
            'refresh': refresh,
            'update_retr': update_retr,
            'expiry': expiry,
            'minimum': minimum,
            'hostmaster': hostmaster,
            'ns_fqdn': ns_fqdn,
        }
        return await self._call(f"/zones/{zoneid}/soa", params, method='post')

    async def get_zone_ns(self, zoneid: int):
        return await self._call(f"/zones/{zoneid}/ns", {})

    async def add_zone_ns(self, zoneid: int, fqdn : str, addrs : list[str]):
        return await self._call(f"/zones/{zoneid}/ns", {'fqdn':fqdn, 'addrs':addrs}, method='post')

    async def del_zone_ns(self, zoneid: int, fqdn : str):
        return await self._call(f"/zones/{zoneid}/ns", {'fqdn':fqdn, 'addrs':[]}, method='delete')

    async def zone_pricelist(self, zoneid: int):
        return await self._call(f"/zones/{zoneid}/pricelist", {})

    async def set_zone_pricelist(self, zoneid : int, operation : str, price : float, valid_from : str):
        return await self._call(f"/zones/{zoneid}/pricelist", {'operation':operation, 'price':price, 'valid_from':valid_from}, method='post')

    async def list_registrars(self, handle : str = None):
        params = {}
        if handle is not None:
            params['handle'] = handle
        return await self._call(f"/registrars", params)

    async def create_registrar(self, handle: str, name: str, system: bool):
        return await self._call(f"/registrars", {'handle':handle, 'name':name, 'system':system}, method='post')

    async def get_registrar_acl(self, regid : int):
        return await self._call(f"/registrars/{regid}/acl", {})

    async def create_registrar_acl(self, regid : int, cert : str, password : str):
        return await self._call(f"/registrars/{regid}/acl", {'cert':cert, 'password':password}, method='post')

    async def update_registrar_acl(self, regid : int, cert : str, password : str):
        return await self._call(f"/registrars/{regid}/acl", {'cert':cert, 'password':password}, method='put')

    async def list_registrar_zones(self, regid : int):
        return await self._call(f"/registrars/{regid}/zones", {})

    async def registrar_add_zone(self, regid : int, fqdn : str):
        return await self._call(f"/registrars/{regid}/zones", [{'zone':fqdn}], method='post')

    async def registrar_del_zone(self, regid : int, fqdn : str):
        return await self._call(f"/registrars/{regid}/zones", [{'zone':fqdn}], method='delete')

async def get_registrar_id(admapi : AdminAPI, handle : str) -> int:
    registrars = await admapi.list_registrars(handle)
    if len(registrars) != 1 or handle is None:
        raise Exception(f"registrar {handle} does not exist")

    return registrars[0]['id']

async def get_zone_id(admapi : AdminAPI, zone : str) -> int:
    zones = await admapi.list_zones(zone)
    if len(zones) != 1 or zone is None:
        raise Exception(f"zone {zone} does not exist")

    return zones[0]['id']

async def main() -> None:
    parser = argparse.ArgumentParser()

    parser.add_argument('--list-registrars', dest='list_registrars', action='store_true', default=False, help="list registrars")
    parser.add_argument('--create-registrar', dest='create_registrar', action='store_true', default=False, help="create registrar: --create-registrar --handle REG --name NAME")
    parser.add_argument('--registrar-acl', dest='registrar_acl', action='store_true', default=False, help="get registrar acl: --regsitrar-acl-get --handle REG")
    parser.add_argument('--set-registrar-acl', dest='set_registrar_acl', action='store_true', default=False, help="set registrar acl: --registrar-acl --handle REG --cert CERT --password PASSWORD")
    parser.add_argument('--registrar-add-zone', dest='registrar_add_zone', action='store_true', default=False, help="registrar add zone: --registrar-add-zone --handle REG --zone fqdn")
    parser.add_argument('--registrar-del-zone', dest='registrar_del_zone', action='store_true', default=False, help="registrar del zone: --registrar-del-zone --handle REG --zone fqdn")
    parser.add_argument('--registrar-list-zones', dest='registrar_list_zones', action='store_true', default=False, help="registrar list zones: --registrar-list-zones --handle REG")

    parser.add_argument('--handle',  type=str, default=None, help="registrar handle")
    parser.add_argument('--cert',  type=str, default=None, help="registrar cert")
    parser.add_argument('--password',  type=str, default=None, help="registrar password")

    # create registrar params
    parser.add_argument('--name',  type=str, default=None, help="registrar name")
    parser.add_argument('--system',  type=bool, default=False, help="if registrar is a system registrar")

    parser.add_argument('--list-zones', dest='list_zone', action='store_true', default=False, help="list zone")
    parser.add_argument('--add-zone', dest='add_zone', action='store_true', default=False, help="add zone: --add-zone --zone fqdn")
    parser.add_argument('--zone-soa', dest='zone_soa', action='store_true', default=False, help="view zone soa: --zone-soa --zone fqdn")
    parser.add_argument('--set-zone-soa', dest='set_zone_soa', action='store_true', default=False, help="set zone soa: --set-zone-soa --zone fqdn --hostmaster HOSTMASTER --ns-fqdn FQDN [--serial SERIAL --ttl TTL --refresh REFRESH] ")
    parser.add_argument('--zone-ns', dest='zone_ns', action='store_true', default=False, help="view zone ns: --zone-ns --zone fqdn")
    parser.add_argument('--add-zone-ns', dest='add_zone_ns', action='store_true', default=False, help="add zone ns: --add-zone-ns --zone fqdn --fqdn FQDN --addrs ADDRS")
    parser.add_argument('--del-zone-ns', dest='del_zone_ns', action='store_true', default=False, help="del zone ns: --del-zone-ns --zone fqdn --fqdn FQDN")
    parser.add_argument('--zone-pricelist', dest='zone_pricelist', action='store_true', default=False, help="list zone price list: --zone-pricelist --zone fqdn")
    parser.add_argument('--zone-set-price', dest='zone_set_price', action='store_true', default=False, help="set zone price: --zone-set-price --zone icasd --operation CreateDomain --price 0")

    parser.add_argument('--zone',  type=str, default=None, help="zone fqdn")
    parser.add_argument('--operation',  type=str, default=None, help="price list operation")
    parser.add_argument('--price',  type=float, default=None, help="operation price")
    parser.add_argument('--valid-from',  type=str, default=str(date.today()), help="start price")

    # soa params
    parser.add_argument('--serial',  type=int, default=None, help="soa serial, if null, current timestamp will be used instead")
    parser.add_argument('--ttl',  type=int, default=14400, help="soa ttl")
    parser.add_argument('--expiry',  type=int, default=2592000, help="soa expiry")
    parser.add_argument('--refresh',  type=int, default=86400, help="soa refresh")
    parser.add_argument('--update_retr',  type=int, default=3600, help="soa refresh")
    parser.add_argument('--minimum',  type=int, default=3600, help="soa refresh")
    parser.add_argument('--hostmaster',  type=str, default=None, help="soa hostmaster")
    parser.add_argument('--ns-fqdn',  type=str, default=None, help="soa ns")

    # ns params
    parser.add_argument('--fqdn',  type=str, default=None, help="soa ns")
    parser.add_argument('--addrs',  type=str, default='', help="soa ns")

    args = parser.parse_args()

    config = read_config()

    admapi = AdminAPI(config['admin_host'])

    if args.list_zone:
        zones = await admapi.list_zones()
        for zone in zones:
            print(zone['fqdn'])

    if args.add_zone:
        zone = await admapi.add_zone(args.zone)
        print(zone)

    if args.zone_soa:
        zone_id = await get_zone_id(admapi, args.zone)
        zone_soa = await admapi.get_zone_soa(zone_id)
        print(zone_soa)

    if args.set_zone_soa:
        zone_id = await get_zone_id(admapi, args.zone)
        zone_soa = await admapi.set_zone_soa(zone_id, args.serial, args.ttl, args.refresh, args.update_retr, args.expiry, args.minimum, args.hostmaster, args.ns_fqdn)
        print(zone_soa)

    if args.zone_ns:
        zone_id = await get_zone_id(admapi, args.zone)
        zone_ns = await admapi.get_zone_ns(zone_id)
        print(zone_ns)

    if args.add_zone_ns:
        zone_id = await get_zone_id(admapi, args.zone)
        zone_ns = await admapi.add_zone_ns(zone_id, args.fqdn, args.addrs.split(','))
        print(zone_ns)

    if args.del_zone_ns:
        zone_id = await get_zone_id(admapi, args.zone)
        zone_ns = await admapi.del_zone_ns(zone_id, args.fqdn)
        print(zone_ns)

    if args.zone_pricelist:
        zone_id = await get_zone_id(admapi, args.zone)
        pricelist = await admapi.zone_pricelist(zone_id)
        for price in pricelist:
            print(price['operation'], price['price'])

    if args.zone_set_price:
        zone_id = await get_zone_id(admapi, args.zone)
        pricelist = await admapi.set_zone_pricelist(zone_id, args.operation, args.price, args.valid_from)
        print(pricelist)

    if args.list_registrars:
        registrars = await admapi.list_registrars()
        for reg in registrars:
            print(reg['handle'])

    if args.create_registrar:
        registrar = await admapi.create_registrar(args.handle, args.name, args.system)
        print(registrar)

    if args.registrar_acl:
        regid = await get_registrar_id(admapi, args.handle)
        reg_acl = await admapi.get_registrar_acl(regid)
        print(reg_acl)

    if args.set_registrar_acl:
        regid = await get_registrar_id(admapi, args.handle)
        reg_acl = await admapi.get_registrar_acl(regid)
        if reg_acl is None or 'detail' in reg_acl:
            reg_acl = await admapi.create_registrar_acl(regid, args.cert, args.password)
        else:
            reg_acl = await admapi.update_registrar_acl(regid, args.cert, args.password)
        print(reg_acl)

    if args.registrar_list_zones:
        regid = await get_registrar_id(admapi, args.handle)
        zones = await admapi.list_registrar_zones(regid)
        for zone in zones:
            print(zone)

    if args.registrar_add_zone:
        regid = await get_registrar_id(admapi, args.handle)
        zone = await admapi.registrar_add_zone(regid, args.zone)
        print(zone)

    if args.registrar_del_zone:
        regid = await get_registrar_id(admapi, args.handle)
        zone = await admapi.registrar_del_zone(regid, args.zone)
        print(zone)

if __name__ == '__main__':
    asyncio.run(main())
