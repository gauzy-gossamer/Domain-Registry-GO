import datetime
from tortoise import Model, fields
from tortoise.contrib.postgres.fields import ArrayField
from fastapi_admin.models import AbstractAdmin


class Admin(AbstractAdmin):
    last_login = fields.DatetimeField(description="Last Login", default=datetime.datetime.now)
    email = fields.CharField(max_length=200, default="")
    intro = fields.TextField(default="")
    created_at = fields.DatetimeField(auto_now_add=True)

    def __str__(self):
        return f"{self.pk}#{self.username}"

class AdminSession(Model):
    admin = fields.ForeignKeyField("models.Admin")
    token = fields.CharField(max_length=255)
    expire = fields.DatetimeField()


class Zone(Model):
    fqdn = fields.CharField(max_length=255)
    ex_period_min = fields.IntField(default=12)
    ex_period_max = fields.IntField(default=12)

    def __str__(self):
        return self.fqdn

class ZoneSOA(Model):
    id = fields.IntField(source_field="zone", pk=True, generated=False)
    zone = fields.ForeignKeyField("models.Zone", source_field="zone")
    ttl = fields.IntField()
    hostmaster = fields.CharField(max_length=255)
    serial = fields.IntField(null=True)
    update_retr = fields.IntField()
    refresh = fields.IntField()
    expiry = fields.IntField()
    minimum = fields.IntField()
    ns_fqdn = fields.CharField(max_length=255)

    class Meta:
        table = 'zone_soa'

class ZoneNS(Model):
    id = fields.IntField(pk=True)
    zone = fields.ForeignKeyField("models.Zone", source_field="zone")
    fqdn = fields.CharField(max_length=255)
    addrs = ArrayField(element_type="inet", null=True, default=[])

    class Meta:
        table = 'zone_ns'

class DomainChecks(Model):
    name = fields.CharField(max_length=255)
    description = fields.CharField(max_length=255)

    class Meta:
        table = 'enum_domain_name_validation_checker'

    def __str__(self):
        return self.description

class DomainCheckMap(Model):
    zone = fields.ForeignKeyField("models.Zone")
    checker = fields.ForeignKeyField("models.DomainChecks")
    
    class Meta:
        table = 'zone_domain_name_validation_checker_map'

class ZoneOperation(Model):
    operation = fields.CharField(max_length=64)

    class Meta:
        table = "enum_operation"

    def __str__(self):
        return self.operation

# fix no timezone error
class DatetimeFieldNoTZ(fields.DateField):
    class _db_postgres:
        SQL_TYPE='TIMESTAMP'

class ZonePrices(Model):
    zone = fields.ForeignKeyField("models.Zone")
    operation = fields.ForeignKeyField("models.ZoneOperation")
    valid_from = DatetimeFieldNoTZ()
    price = fields.DecimalField(12, decimal_places=2)

    class Meta:
        table = "price_list"

class Registrar(Model):
    object_id = fields.IntField()
    handle = fields.CharField(max_length=255)
    name = fields.CharField(max_length=255)
    intpostal = fields.CharField(max_length=512)
    intaddress = fields.JSONField()
    locpostal = fields.CharField(max_length=512)
    locaddress = fields.JSONField()
    legaladdress = fields.JSONField()
    taxpayernumbers = fields.CharField(max_length=32, null=True)
    telephone = fields.JSONField()
    fax = fields.JSONField()

    email = fields.JSONField()
    notify_email = fields.JSONField()
    info_email = fields.JSONField()
    url = fields.CharField(max_length=1024)
    www = fields.CharField(max_length=255)
    system = fields.BooleanField()

    def __str__(self):
        return self.handle

class RegistrarACL(Model):
    registrarid = fields.ForeignKeyField("models.Registrar", source_field="registrarid")
    cert = fields.CharField(max_length=1024)
    password = fields.CharField(max_length=64)

class RegistrarIPAddr(Model):
    registrarid = fields.ForeignKeyField("models.Registrar", source_field="registrarid")
    ipaddr = fields.CharField(max_length=128)

    class Meta:
        table = 'registrar_ipaddr_map'

class RegistrarBalance(Model):
    registrar = fields.ForeignKeyField("models.Registrar")
    zone = fields.ForeignKeyField("models.Zone")
    credit = fields.DecimalField(12, decimal_places=2)

    class Meta:
        table = 'registrar_credit'

class RegistrarTransaction(Model):
    registrar_credit = fields.ForeignKeyField("models.RegistrarBalance")
    balance_change = fields.DecimalField(12, decimal_places=2)

    class Meta:
        table = 'registrar_credit_transaction'

class RegistrarInvoice(Model):
    registrarid = fields.ForeignKeyField("models.Registrar", source_field="registrarid")
    zone = fields.ForeignKeyField("models.Zone", source_field='zone')
    fromdate = fields.DateField(auto_now_add=True)

    class Meta:
        table = 'registrarinvoice'

class Invoice(Model):
    registrar = fields.ForeignKeyField("models.Registrar")
    zone = fields.ForeignKeyField("models.Zone")
    crdate = fields.DatetimeField()
    balance = fields.DecimalField(12, decimal_places=2)
    comment = fields.TextField()

    class Meta:
        table = 'invoice'
