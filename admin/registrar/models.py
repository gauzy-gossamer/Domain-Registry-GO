import sqlalchemy
from sqlalchemy.dialects.postgresql import UUID, JSONB, INET

metadata = sqlalchemy.MetaData()

registrars_table = sqlalchemy.Table(
    "registrar",
    metadata,
    sqlalchemy.Column("id", sqlalchemy.Integer, primary_key=True),
    sqlalchemy.Column("object_id", sqlalchemy.Integer),
    sqlalchemy.Column("handle", sqlalchemy.String(255), unique=True),
    sqlalchemy.Column("name", sqlalchemy.String(255)),
    sqlalchemy.Column("intpostal", sqlalchemy.String(512)),
    sqlalchemy.Column("intaddress", JSONB),
    sqlalchemy.Column("locpostal", sqlalchemy.String(512)),
    sqlalchemy.Column("locaddress", JSONB),
    sqlalchemy.Column("legaladdress", JSONB),
    sqlalchemy.Column("taxpayernumbers", sqlalchemy.String(32)),
    sqlalchemy.Column("telephone", JSONB),
    sqlalchemy.Column("fax", JSONB),
    sqlalchemy.Column("email", JSONB),
    sqlalchemy.Column("notify_email", JSONB),
    sqlalchemy.Column("info_email", JSONB),
    sqlalchemy.Column("url", sqlalchemy.String(1024)),
    sqlalchemy.Column("www", sqlalchemy.String(255)),
    sqlalchemy.Column("system", sqlalchemy.Boolean),
    sqlalchemy.Column("epp_requests_limit", sqlalchemy.Integer),
)

registraracl_table = sqlalchemy.Table(
    "registraracl",
    metadata,
    sqlalchemy.Column("id", sqlalchemy.Integer, primary_key=True),
    sqlalchemy.Column("registrarid", sqlalchemy.Integer),
    sqlalchemy.Column("cert", sqlalchemy.String(1024)),
    sqlalchemy.Column("password", sqlalchemy.String(64)),
)

registrar_ipaddr_table = sqlalchemy.Table(
    "registrar_ipaddr_map",
    metadata,
    sqlalchemy.Column("id", sqlalchemy.Integer, primary_key=True),
    sqlalchemy.Column("registrarid", sqlalchemy.Integer),
    sqlalchemy.Column("ipaddr", INET),
)

registrar_invoice_table = sqlalchemy.Table(
    "registrarinvoice",
    metadata,
    sqlalchemy.Column("registrarid", sqlalchemy.Integer),
    sqlalchemy.Column("zone", sqlalchemy.Integer),
    sqlalchemy.Column("fromdate", sqlalchemy.DateTime),
    sqlalchemy.Column("todate", sqlalchemy.DateTime),
)