import sqlalchemy
from sqlalchemy.dialects.postgresql import UUID, INET, ARRAY

metadata = sqlalchemy.MetaData()

zones_table = sqlalchemy.Table(
    "zone",
    metadata,
    sqlalchemy.Column("id", sqlalchemy.Integer, primary_key=True),
    sqlalchemy.Column("fqdn", sqlalchemy.String(255), unique=True, index=True),
    sqlalchemy.Column("ex_period_min", sqlalchemy.Integer),
    sqlalchemy.Column("ex_period_max", sqlalchemy.Integer),
)

zone_groups_table = sqlalchemy.Table(
    "zone_group",
    metadata,
    sqlalchemy.Column("zone_id", sqlalchemy.Integer, index=True),
    sqlalchemy.Column("group_id", sqlalchemy.Integer),
)

zone_soa_table = sqlalchemy.Table(
    "zone_soa",
    metadata,
    sqlalchemy.Column("zone", sqlalchemy.Integer),
    sqlalchemy.Column("ttl", sqlalchemy.Integer),
    sqlalchemy.Column("hostmaster", sqlalchemy.String(255)),
    sqlalchemy.Column("serial", sqlalchemy.Integer),
    sqlalchemy.Column("update_retr", sqlalchemy.Integer),
    sqlalchemy.Column("refresh", sqlalchemy.Integer),
    sqlalchemy.Column("expiry", sqlalchemy.Integer),
    sqlalchemy.Column("minimum", sqlalchemy.Integer),
    sqlalchemy.Column("ns_fqdn", sqlalchemy.String(255)),
)

zone_ns_table = sqlalchemy.Table(
    "zone_ns",
    metadata,
    sqlalchemy.Column("id", sqlalchemy.Integer),
    sqlalchemy.Column("zone", sqlalchemy.Integer),
    sqlalchemy.Column("fqdn", sqlalchemy.String(255)),
    sqlalchemy.Column("addrs", ARRAY(INET)),
)

price_list_table = sqlalchemy.Table(
    "price_list",
    metadata,
    sqlalchemy.Column("id", sqlalchemy.Integer, primary_key=True),
    sqlalchemy.Column("zone_id", sqlalchemy.Integer),
    sqlalchemy.Column("operation_id", sqlalchemy.Integer),
    sqlalchemy.Column("valid_from", sqlalchemy.DateTime),
    sqlalchemy.Column("valid_to", sqlalchemy.DateTime),
    sqlalchemy.Column("price", sqlalchemy.Numeric),
)

enum_operation_table = sqlalchemy.Table(
    "enum_operation",
    metadata,
    sqlalchemy.Column("id", sqlalchemy.Integer, primary_key=True),
    sqlalchemy.Column("operation", sqlalchemy.String(64)),
)
