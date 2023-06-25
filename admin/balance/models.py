import sqlalchemy

metadata = sqlalchemy.MetaData()

credit_table = sqlalchemy.Table(
    "registrar_credit",
    metadata,
    sqlalchemy.Column("id", sqlalchemy.Integer, primary_key=True),
    sqlalchemy.Column("credit", sqlalchemy.Numeric),
    sqlalchemy.Column("registrar_id", sqlalchemy.Integer),
    sqlalchemy.Column("zone_id", sqlalchemy.Integer),
)

credit_transaction_table = sqlalchemy.Table(
    "registrar_credit_transaction",
    metadata,
    sqlalchemy.Column("id", sqlalchemy.Integer, primary_key=True),
    sqlalchemy.Column("balance_change", sqlalchemy.Numeric),
    sqlalchemy.Column("registrar_credit_id", sqlalchemy.Integer),
)

invoice_table = sqlalchemy.Table(
    "invoice",
    metadata,
    sqlalchemy.Column("id", sqlalchemy.Integer, primary_key=True),
    sqlalchemy.Column("zone_id", sqlalchemy.Integer),
    sqlalchemy.Column("crdate", sqlalchemy.DateTime),
    sqlalchemy.Column("taxdate", sqlalchemy.Date),
    sqlalchemy.Column("prefix", sqlalchemy.Integer),
    sqlalchemy.Column("registrar_id", sqlalchemy.Integer),
    sqlalchemy.Column("balance", sqlalchemy.Numeric),
    sqlalchemy.Column("vat", sqlalchemy.Numeric),
    sqlalchemy.Column("invoice_prefix_id", sqlalchemy.Integer),
    sqlalchemy.Column("comment", sqlalchemy.String(255)),
)

invoice_prefix_table = sqlalchemy.Table(
    "invoice_prefix",
    metadata,
    sqlalchemy.Column("id", sqlalchemy.Integer, primary_key=True),
    sqlalchemy.Column("zone_id", sqlalchemy.Integer),
    sqlalchemy.Column("prefix", sqlalchemy.Integer),
    sqlalchemy.Column("typ", sqlalchemy.Integer),
    sqlalchemy.Column("year", sqlalchemy.Numeric),
)
