from models.database import database

import zone.schema as schema
from zone.models import zones_table, zone_soa_table, zone_ns_table, price_list_table, enum_operation_table

async def create_zone(zone : schema.Zone) -> dict:
    query = zones_table.insert().values(
        fqdn=zone.fqdn,
        ex_period_min=zone.ex_period_min,
        ex_period_max=zone.ex_period_max,
    )

    zone_id = await database.execute(query)

    return {**zone.dict(), "id":zone_id}

async def get_zone(zone_id : int):
    query = zones_table.select().where(zones_table.c.id==zone_id)
    return await database.fetch_one(query)

async def get_zones(zone : str = None):
    query = zones_table.select()
    if zone is not None:
        query = query.filter(zones_table.c.fqdn == zone)
    return await database.fetch_all(query)

async def get_zone_pricelist(zone_id : int): 
    query = zones_table.select().with_only_columns(
                    zones_table.c.id, zones_table.c.fqdn, enum_operation_table.c.operation, price_list_table.c.valid_from, price_list_table.c.price
                ).join(price_list_table, price_list_table.c.zone_id == zones_table.c.id
                ).join(enum_operation_table, enum_operation_table.c.id == price_list_table.c.operation_id
                ).where(zones_table.c.id == zone_id)
    return await database.fetch_all(query)

async def add_zone_pricelist(zone_id : int, zone_price : schema.ZonePriceList): 
    query = enum_operation_table.select().where(enum_operation_table.c.operation == zone_price.operation).with_only_columns(enum_operation_table.c.id)
    op = await database.fetch_one(query)
    if op is None:
        raise Exception(f"operation {zone_price.operation} does not exist")

    query = price_list_table.select().filter(price_list_table.c.zone_id == zone_id, price_list_table.c.operation_id == op.id, price_list_table.c.valid_to == None)
    price_val = await database.fetch_one(query)

    if price_val is None:
        query = price_list_table.insert().values(
            zone_id = zone_id,
            price = zone_price.price,
            operation_id = op.id,
            valid_from = zone_price.valid_from,
        )
    else:
        query = price_list_table.update().values(
            price = zone_price.price,
        ).where(price_list_table.c.id == price_val.id)

    await database.execute(query)

    return {**zone_price.dict(), "zoneid":zone_id}

async def get_zone_soa(zone_id : int):
    query = zone_soa_table.select().where(zone_soa_table.c.zone == zone_id)
    return await database.fetch_one(query)

async def update_zone_soa(zone_id : int, zone_soa : schema.ZoneSoa):
    query = zone_soa_table.select().where(zone_soa_table.c.zone == zone_id)
    zone_soa_obj = await database.fetch_one(query)
    if zone_soa_obj is None:
        query = zone_soa_table.insert().values(
            zone = zone_id,
            ttl = zone_soa.ttl,
            serial = zone_soa.serial,
            refresh = zone_soa.refresh,
            update_retr = zone_soa.update_retr,
            expiry = zone_soa.expiry,
            minimum = zone_soa.minimum,
            hostmaster = zone_soa.hostmaster,
            ns_fqdn = zone_soa.ns_fqdn,
        )
    else:
        query = zone_soa_table.update().values(
            ttl = zone_soa.ttl,
            serial = zone_soa.serial,
            refresh = zone_soa.refresh,
            update_retr = zone_soa.update_retr,
            expiry = zone_soa.expiry,
            minimum = zone_soa.minimum,
            hostmaster = zone_soa.hostmaster,
            ns_fqdn = zone_soa.ns_fqdn,
        ).where(zone_soa_table.c.zone == zone_id)

    await database.execute(query)

    return {**zone_soa.dict()}

async def get_zone_ns(zone_id : int):
    query = zone_ns_table.select().where(zone_ns_table.c.zone == zone_id)
    return await database.fetch_all(query)

async def add_zone_ns(zone_id : int, zone_ns : schema.ZoneNs):
    query = zone_ns_table.select().where(zone_ns_table.c.id == zone_id, zone_ns_table.c.fqdn == zone_ns.fqdn)
    zone_ns_obj = await database.fetch_one(query)
    if zone_ns_obj is None:
        query = zone_ns_table.insert().values(
            zone = zone_id,
            fqdn = zone_ns.fqdn,
            addrs = zone_ns.addrs,
        )
    else:
        query = zone_ns_table.update().values(
            fqdn = zone_ns.fqdn,
            addrs = zone_ns.addrs,
        ).where(zone_ns_table.c.id == zone_ns_obj.id)

    await database.execute(query)

    return await get_zone_ns(zone_id)

async def del_zone_ns(zone_id : int, zone_ns : schema.ZoneNs):
    query = zone_ns_table.delete().where(zone_ns_table.c.zone == zone_id, zone_ns_table.c.fqdn == zone_ns.fqdn)

    await database.execute(query)

    return await get_zone_ns(zone_id)
