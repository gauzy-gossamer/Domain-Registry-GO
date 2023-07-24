from starlette.requests import Request

from fastapi_admin.app import app
from fastapi_admin.enums import Method
from fastapi_admin.resources import Action, Dropdown, Field, Link, Model, ToolbarAction
from fastapi_admin.widgets import displays, filters, inputs

import panel.fields as p_inputs
from panel.models import Admin, Zone, ZoneNS, ZoneSOA, ZonePrices, ZoneOperation, DomainChecks, DomainCheckMap, Registrar, RegistrarACL, RegistrarIPAddr, RegistrarBalance, RegistrarInvoice, RegistrarTransaction, Invoice


@app.register
class Dashboard(Link):
    label = "Dashboard"
    icon = "fas fa-home"
    url = "/admin"


@app.register
class AdminResource(Model):
    label = "Admin"
    model = Admin
    icon = "fas fa-user"
    page_title = "Admin list"
    filters = [
        filters.Search(
            name="username",
            label="Name",
            search_mode="contains",
            placeholder="Search for username",
        ),
        filters.Date(name="created_at", label="CreatedAt"),
    ]
    fields = [
        "id",
        "username",
        Field(name="password", label="Password", display=displays.InputOnly(), input_=inputs.Password()),
        Field(name="email", label="Email", input_=inputs.Email()),
        "created_at",
    ]

    async def get_actions(self, request: Request) -> list[Action]:
        return [
            Action(
                label=_("update"), icon="ti ti-edit", name="update", method=Method.GET, ajax=False
            ),
        ]

    async def get_bulk_actions(self, request: Request) -> list[Action]:
        return []

@app.register
class Registrators(Dropdown):
    class RegistrarResource(Model):
        label = "Registrar"
        page_title = "Registrars"
        model = Registrar
        filters = [
            filters.Search(name="handle", label="Registrar", search_mode="contains", placeholder="Search for registrar")
        ]
        fields = ["handle", "name", 
            Field("intpostal", input_=inputs.Text(help_text='international organization name')),
            Field("intaddress", label='International address', display=displays.InputOnly(), input_=inputs.Json(help_text='international address', null=True)),
            Field("locpostal", display=displays.InputOnly(), input_=inputs.Text(help_text='localized organization name', null=True)),
            Field("locaddress", label='Local address', display=displays.InputOnly(), input_=inputs.Json(help_text='localized address', null=True)),
            Field("legaladdress", label='Legal address', display=displays.InputOnly(), input_=inputs.Json(null=True)),
            Field("telephone", display=displays.InputOnly(), input_=inputs.Json(null=True)),
            Field("fax", display=displays.InputOnly(), input_=inputs.Json(null=True)),
            Field("taxpayernumbers", label='taxpayer numbers', display=displays.InputOnly(), input_=inputs.Text(null=True)),
            "email",
            Field("www", label='WWW', display=displays.InputOnly(), input_=inputs.Text(null=True)),
            Field("url", label='URL', display=displays.InputOnly(), input_=inputs.Text(null=True)),
            "system"
        ]

        async def get_actions(self, request: Request) -> list[Action]:
            return [
                Action(label=_("update"), icon="ti ti-edit", name="update", method=Method.GET, ajax=False),
            ]

        async def get_bulk_actions(self, request: Request) -> list[Action]:
            return []

    class RegistrarACLResource(Model):
        label = "Registrar ACL"
        page_title = "Registrar access parameters"
        page_pre_title = "Registrar certificates and password"
        model = RegistrarACL
        filters = [
            filters.Search(name="registrar__handle", label="Registrar", search_mode="contains", placeholder="Search for registrar")
        ]
        fields = [
            Field(name="registrarid", label='registrar', input_=inputs.ForeignKey(model=Registrar)), 
            Field("cert", label='certificate', input_=inputs.Text(placeholder="05:3D:AE:11:5C:CC:A6:C1:08:61:24:E1:EA:12:67:50", help_text='certificate md5 checksum')),
            "password",
        ]

    class RegistrarIPAddrResource(Model):
        label = "Registrar IPs"
        page_title = "Registrar IP-Addresses"
        page_pre_title = "Only access from these addresses will be allowed"
        model = RegistrarIPAddr
        filters = [
            filters.Search(name="registrar__handle", label="Registrar", search_mode="contains", placeholder="Search for registrar")
        ]
        fields = [
            Field(name="registrarid", label='Registrar', input_=inputs.ForeignKey(model=Registrar)), 
            Field("ipaddr", label='IP-address', input_=p_inputs.IPAddress())
        ]

    class RegistrarZoneAccessResource(Model):
        label = "Registrar zones"
        page_title = "Registrar zones"
        page_pre_title = "Zones that registrars have access to"
        model = RegistrarInvoice
        filters = [
            filters.Search(
                name="registrar__handle",
                label="Registrar",
                search_mode="contains",
                placeholder="Search for registrar",
            )
        ]

        async def get_actions(self, request: Request) -> list[Action]:
            return [Action(label=_("delete"), icon="ti ti-trash", name="delete", method=Method.DELETE)]

        fields = [Field(name="registrarid", label='Registrar', input_=inputs.ForeignKey(model=Registrar)), Field(name="zone", input_=inputs.ForeignKey(model=Zone)), "fromdate"]

    label = "Registrators"
    icon = "fas fa-briefcase"
    resources = [RegistrarResource, RegistrarACLResource, RegistrarIPAddrResource, RegistrarZoneAccessResource]


class RegistrarTransactionResource(Model):
    label = "Add balance"
    model = RegistrarTransaction
    fields = [
        Field(name="registrar_credit", label='Registrar', input_=p_inputs.ForeignKeyRelated(model=RegistrarBalance, related_fields=["registrar__handle", "zone__fqdn"])), 
        Field("balance_change", label='Balance change'),
    ]

@app.register
class Balance(Dropdown):
    class RegistrarBalanceResource(Model):
        label = "Registrar balance"
        model = RegistrarBalance
        fields = [Field(name="registrar", input_=inputs.ForeignKey(model=Registrar)), Field(name="zone", input_=inputs.ForeignKey(model=Zone)), "credit"]

        filters = [
            filters.Search(name="registrar__handle", label="Registrar", search_mode="contains", placeholder="Search for registrar")
        ]

        async def get_toolbar_actions(self, request: Request) -> list[ToolbarAction]:
            return [
                ToolbarAction(label=_("Add Balance"), icon="fas fa-plus", name="add_balance", method=Method.GET, ajax=False, class_='btn-dark'),
                ToolbarAction(label=_("create"), icon="fas fa-plus", name="create", method=Method.GET, ajax=False, class_='btn-dark'),
            ]

        async def get_actions(self, request: Request) -> list[Action]:
            return []

        async def get_bulk_actions(self, request: Request) -> list[Action]:
            return []

    class InvoiceResource(Model):
        label = "Registrar payments"
        model = Invoice
        fields = [
            Field(name="registrar", input_=inputs.ForeignKey(model=Registrar)), 
            Field(name="zone", input_=inputs.ForeignKey(model=Zone)), 
            Field("crdate", label="payment created"),
            Field("balance", label="payment sum"),
            "comment",
        ]

        filters = [
            filters.Search(name="registrar__handle", label="Registrar", search_mode="contains", placeholder="Search for registrar")
        ]

        async def get_toolbar_actions(self, request: Request) -> list[ToolbarAction]:
            return []

        async def get_actions(self, request: Request) -> list[Action]:
            return []

        async def get_bulk_actions(self, request: Request) -> list[Action]:
            return []

    label = "Balance"
    icon = "fas fa-dollar-sign"
    resources = [RegistrarBalanceResource, InvoiceResource]

@app.register
class Zones(Dropdown):
    class ZoneResource(Model):
        label = "Zone"
        page_title = "Zone list"
        model = Zone
        fields = [Field("fqdn", label='Zone', input_=p_inputs.Hostname(placeholder='example.com'))]

    class ZoneSOAResource(Model):
        label = "Zone SOA"
        page_title = "Zone SOA record"
        model = ZoneSOA
        filters = [
            filters.Search(name="zone__fqdn", label="Zone", search_mode="contains", placeholder="Search for zone")
        ]
        fields = [
            p_inputs.GetZoneName(name="tzone", label='zone name'),
            Field("ttl", label='TTL', display=displays.InputOnly(), input_=p_inputs.Num(placeholder="14400")),
            Field("serial", label="Zone serial", input_=p_inputs.Num(help_text="leave empty to use timestamp as serial", null=True)),
            Field("update_retr", display=displays.InputOnly(), input_=p_inputs.Num()),
            Field("refresh", display=displays.InputOnly(), input_=p_inputs.Num()),
            Field("expiry", display=displays.InputOnly(), input_=p_inputs.Num()),
            Field("minimum", display=displays.InputOnly(), input_=p_inputs.Num()),
            Field("hostmaster", input_=p_inputs.Hostname(placeholder="hostmaster.example.com")),
            Field("ns_fqdn", label='Nameserver', input_=p_inputs.Hostname(placeholder="ns1.example.com"))
        ]

        @classmethod
        def get_fields(self, is_display: bool = True):
            ret = super().get_fields(is_display=is_display)
            # override default behaviour, since it doesn't show primary keys
            if not is_display:
                ret[0] = Field(name="zone", input_=inputs.ForeignKey(model=Zone))
            return ret 

        @classmethod
        async def resolve_data(cls, request: Request, data):
            ret, m2m_ret = await super().resolve_data(request, data)
            if 'serial' in ret and ret['serial'].strip() == '':
                ret['serial'] = None
            ret['id'] = ret['zone'].id

            return ret, m2m_ret

        async def get_bulk_actions(self, request: Request) -> list[Action]:
            return []

    class ZoneNSResource(Model):
        label = "Zone Nameservers"
        page_title = "Zone Nameservers"
        page_pre_title = "Authoritative nameservers for zones"
        model = ZoneNS
        filters = [
            filters.Search(name="zone__fqdn", label="Zone", search_mode="contains", placeholder="Search for zone")
        ]
        fields = [
            Field("zone", input_=inputs.ForeignKey(model=Zone)), 
            Field("fqdn", label='Nameserver', input_=p_inputs.Hostname(placeholder="ns1.example.com")),
#            Field("addrs", label='IP-Addresses', input_=inputs.Text(placeholder="10.10.0.1")),
        ]

    class ZoneDomainNameChecks(Model):
        label = "Domain name checks"
        page_title = "Domain name checks"
        page_pre_title = "Domain name validators per zone"
        model = DomainCheckMap
        filters = [
            filters.Search(name="zone__fqdn", label="Zone", search_mode="contains", placeholder="Search for zone")
        ]
        fields = [Field(name="zone", input_=inputs.ForeignKey(model=Zone)), Field(name="checker", input_=inputs.ForeignKey(model=DomainChecks))]

    class ZonePricesResource(Model):
        label = "Zone prices"
        page_title = "Zone operation prices"
        model = ZonePrices
        filters = [
            filters.Search(name="zone__fqdn", label="Zone", search_mode="contains", placeholder="Search for zone")
        ]
        fields = [
            Field("zone", input_=inputs.ForeignKey(model=Zone)), 
            Field("operation", input_=inputs.ForeignKey(model=ZoneOperation)), 
            Field("valid_from", label="Valid from", input_=inputs.DateTime()),
            "price",
        ]

    label = "Zones"
    icon = "fas fa-cog"
    resources = [ZoneResource, ZoneSOAResource, ZoneNSResource, ZoneDomainNameChecks, ZonePricesResource]
