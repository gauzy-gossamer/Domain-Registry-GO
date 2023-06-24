from datetime import date
from models.database import database

import registrar.schema as schema
from registrar.models import registrars_table, registraracl_table, registrar_ipaddr_table, registrar_invoice_table
from zone.models import zones_table

# create object in object_registry table
async def create_object(database, handle : str, registrar : int, object_type : str) -> int:
    object_type_id = await database.execute("SELECT get_object_type_id(:val)", values={'val':object_type})

    object_id = await database.execute("SELECT create_object(:registrar, :handle, :object_type_id)",
                    values={'registrar':registrar, 'handle':handle, 'object_type_id':object_type_id})
 
    await database.execute("INSERT INTO object(id, clid) VALUES(:object_id, :registrar)",
                    values={'object_id':object_id, 'registrar':registrar})

    return object_id

async def test_system_reg() -> bool:
    query = registrars_table.select().where(registrars_table.c.system==True)
    ret = await database.fetch_one(query)
    return ret is not None

async def create_registrar(registrar : schema.Registrar) -> dict:
    # system registrar is unique
    if registrar.system and await test_system_reg():
        raise Exception("system registrar already exists")

    async with database.transaction():
        query = registrars_table.insert().values(
            handle=registrar.handle,
            system=registrar.system,
        )

        registrar_id = await database.execute(query)

        object_id = await create_object(database, registrar.handle, registrar_id, "registrar")

        # connect registrar table to object_registry
        query = registrars_table.update().values(object_id=object_id).where(registrars_table.c.id == registrar_id)
        await database.execute(query)

    return {**registrar.dict(), "id":registrar_id, "object_id":object_id}

async def get_registrar(reg_id : int):
    query = registrars_table.select().where(registrars_table.c.id == reg_id)
    return await database.fetch_one(query)

async def get_registrars(handle: str = None):
    query = registrars_table.select()
    if handle is not None:
        query = query.filter(registrars_table.c.handle == handle)
    return await database.fetch_all(query)

async def get_registraracl(reg_id : int):
    query = registraracl_table.select().where(registraracl_table.c.registrarid == reg_id)
    return await database.fetch_one(query)

async def create_registraracl(reg_id : int, regacl : schema.RegistrarAclCreate) -> None:
    query = registraracl_table.select().where(registraracl_table.c.registrarid == reg_id)
    regobj = await database.fetch_one(query)
    if regobj:
        raise Exception("registrar acl already exists")

    query = registraracl_table.insert().values(
        cert=regacl.cert,
        password=regacl.password,
        registrarid=reg_id,
    )

    acl_id = await database.execute(query)

    return {**regacl.dict(), "id":acl_id, "registrarid":reg_id}

async def update_registraracl(reg_id : int, regacl : schema.RegistrarAclCreate) -> None:
    query = registraracl_table.select().where(registraracl_table.c.registrarid == reg_id)
    regobj = await database.fetch_one(query)
    if not regobj:
        raise Exception("registrar acl does not exist")

    query = registraracl_table.update().values(
        cert=regacl.cert,
        password=regacl.password,
    ).where(registraracl_table.c.registrarid == reg_id)

    await database.execute(query)
    return {**regacl.dict(), "id":regobj.id, "registrarid":reg_id}

async def get_registrar_ips(reg_id : int):
    query = registrar_ipaddr_table.select().where(registrar_ipaddr_table.c.registrarid == reg_id)
    return await database.fetch_all(query)

async def update_registrar_ips(reg_id : int, ips : list[schema.RegistrarIpAddrBase]):
    query = registrar_ipaddr_table.select().where(registrar_ipaddr_table.c.registrarid == reg_id)
    ips_obj = await database.fetch_all(query)

    present_ips = {obj.ipaddr:obj.id for obj in ips_obj}
    new_ips = {obj.ipaddr for obj in ips}
    async with database.transaction():
        for ip in new_ips:
            if ip in present_ips:
                continue
            query = registrar_ipaddr_table.insert().values(
                ipaddr=ip,
                registrarid=reg_id,
            )
            await database.execute(query)

        for ip in present_ips:
            if ip in new_ips:
                continue
            query = registrar_ipaddr_table.delete().where(
                registrar_ipaddr_table.c.id == present_ips[ip],
            )
            await database.execute(query)

    return ips

async def get_registrar_zones(reg_id : int):
    query = registrar_invoice_table.select().with_only_columns(
        zones_table.c.fqdn.label("zone"), registrar_invoice_table.c.fromdate
    ).join(zones_table, zones_table.c.id == registrar_invoice_table.c.zone
    ).where(registrar_invoice_table.c.registrarid == reg_id, registrar_invoice_table.c.todate == None)
    return await database.fetch_all(query)

async def add_registrar_zones(reg_id : int, reg_zones : list[schema.RegistrarZones]):
    query = registrar_invoice_table.select().with_only_columns(
        registrar_invoice_table.c.zone
    ).where(registrar_invoice_table.c.registrarid == reg_id, registrar_invoice_table.c.todate == None)
    reg_zone_objs = await database.fetch_all(query)
    present_zones = {obj.zone for obj in reg_zone_objs}

    async with database.transaction():
        for zone in reg_zones:
            query = zones_table.select().where(zones_table.c.fqdn == zone.zone)
            zone_obj = await database.fetch_one(query)
            if zone_obj is None:
                raise Exception(f"zone {zone.zone} does not exist")
            if zone_obj.id in present_zones:
                continue

            query = registrar_invoice_table.insert().values(
                registrarid=reg_id,
                zone=zone_obj.id,
                fromdate=zone.fromdate if zone.fromdate is not None else date.today()
            )
            await database.execute(query)

    return await get_registrar_zones(reg_id)

async def del_registrar_zones(reg_id : int, reg_zones : list[schema.RegistrarZones]):
    query = registrar_invoice_table.select().with_only_columns(
        registrar_invoice_table.c.zone
    ).where(registrar_invoice_table.c.registrarid == reg_id, registrar_invoice_table.c.todate == None)
    reg_zone_objs = await database.fetch_all(query)
    present_zones = {obj.zone for obj in reg_zone_objs}

    async with database.transaction():
        for zone in reg_zones:
            query = zones_table.select().where(zones_table.c.fqdn == zone.zone)
            zone_obj = await database.fetch_one(query)
            if zone_obj is None:
                raise Exception(f"zone {zone.zone} does not exist")
            if zone_obj.id not in present_zones:
                continue

            query = registrar_invoice_table.update().values(
                todate=date.today()
            ).where(registrar_invoice_table.c.registrarid==reg_id, registrar_invoice_table.c.zone==zone_obj.id)
            await database.execute(query)

    return await get_registrar_zones(reg_id)
