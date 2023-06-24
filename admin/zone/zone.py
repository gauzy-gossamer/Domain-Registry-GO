from models.database import database

import zone.schema as schema
from zone.models import zones_table, price_list_table, enum_operation_table

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
