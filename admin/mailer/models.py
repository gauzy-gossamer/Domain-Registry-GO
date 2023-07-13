import sqlalchemy

metadata = sqlalchemy.MetaData()

mail_request_table = sqlalchemy.Table(
    "mail_request",
    metadata,
    sqlalchemy.Column("id", sqlalchemy.Integer, primary_key=True),
    sqlalchemy.Column("object_id", sqlalchemy.Integer),
    sqlalchemy.Column("domain_id", sqlalchemy.Integer),
    sqlalchemy.Column("requested", sqlalchemy.DateTime),
    sqlalchemy.Column("request_type_id", sqlalchemy.Integer),
    sqlalchemy.Column("sent_mail", sqlalchemy.Boolean),
    sqlalchemy.Column("tries", sqlalchemy.Integer),
    sqlalchemy.Column("mail_error", sqlalchemy.String(255)),
)

mail_request_type_table = sqlalchemy.Table(
    "mail_request_type",
    metadata,
    sqlalchemy.Column("id", sqlalchemy.Integer, primary_key=True),
    sqlalchemy.Column("request_type", sqlalchemy.String(255)),
    sqlalchemy.Column("subject", sqlalchemy.String(255)),
    sqlalchemy.Column("template", sqlalchemy.Text),
    sqlalchemy.Column("active", sqlalchemy.Boolean),
)
