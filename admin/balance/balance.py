from models.database import database

import balance.schema as schema
from balance.models import credit_table, credit_transaction_table, invoice_table
from registrar.models import registrars_table
from zone.models import zones_table

async def get_registrar_balance(reg_id: int = None):
    query = credit_table.select().with_only_columns(
        registrars_table.c.handle, credit_table.c.credit, zones_table.c.fqdn.label("zone")
    ).join(zones_table, zones_table.c.id == credit_table.c.zone_id
    ).join(registrars_table, registrars_table.c.id == credit_table.c.registrar_id)
    if reg_id is not None:
        query = query.where(credit_table.c.registrar_id == reg_id)
    return await database.fetch_all(query)

async def add_registrar_balance(reg_id: int, add_credit: schema.AddRegistrarCredit):
    query = zones_table.select().where(zones_table.c.fqdn == add_credit.zone)
    zone_obj = await database.fetch_one(query)
    if zone_obj is None:
        raise Exception(f"zone {add_credit.zone} does not exist")

    query = credit_table.select().where(credit_table.c.zone_id == zone_obj.id, credit_table.c.registrar_id == reg_id)
    credit_obj = await database.fetch_one(query)

    if credit_obj is None:
        query = credit_table.insert().values(
            zone_id = zone_obj.id, 
            registrar_id = reg_id,
        )
        registrar_credit_id = await database.execute(query)
    else:
        registrar_credit_id = credit_obj.id

    query = credit_transaction_table.insert().values( 
        registrar_credit_id = registrar_credit_id,
        balance_change = add_credit.balance_change,
    )

    await database.execute(query)

    return await get_registrar_balance(reg_id)

async def get_registrar_invoice(reg_id: int = None):
    query = invoice_table.select().with_only_columns(
        registrars_table.c.handle, invoice_table.c.balance, zones_table.c.fqdn.label("zone"), invoice_table.c.crdate, invoice_table.c.comment
    ).join(zones_table, zones_table.c.id == invoice_table.c.zone_id
    ).join(registrars_table, registrars_table.c.id == invoice_table.c.registrar_id)
    if reg_id is not None:
        query = query.where(invoice_table.c.registrar_id == reg_id)
    return await database.fetch_all(query)
