from starlette.requests import Request

from fastapi_admin.app import app
from fastapi_admin.enums import Method
from fastapi_admin.resources import Action, Dropdown, Field, Link, Model, ToolbarAction
from fastapi_admin.widgets import displays, filters, inputs

from panel.models import Admin, Zone, ZoneSOA, DomainChecks, DomainCheckMap, Registrar, RegistrarACL, RegistrarIPAddr, RegistrarBalance, RegistrarInvoice


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
    page_pre_title = "admin list"
    page_title = "admin model"
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
        Field(
            name="password",
            label="Password",
            display=displays.InputOnly(),
            input_=inputs.Password(),
        ),
        Field(name="email", label="Email", input_=inputs.Email()),
        "created_at",
    ]

    async def get_actions(self, request: Request) -> list[Action]:
        return []

    async def get_bulk_actions(self, request: Request) -> list[Action]:
        return []

@app.register
class Registrators(Dropdown):
    class RegistrarResource(Model):
        label = "Registrar"
        model = Registrar
        fields = ["handle", "name", "email", "www", "url", "system"]

    class RegistrarACLResource(Model):
        label = "Registrar ACL"
        model = RegistrarACL
        fields = [
            Field(name="registrarid", input_=inputs.ForeignKey(model=Registrar)), 
            Field("cert", label='certificate', input_=inputs.Text(placeholder="05:3D:AE:11:5C:CC:A6:C1:08:61:24:E1:EA:12:67:50", help_text='certificate md5 checksum')),
            "password",
        ]

    class RegistrarIPAddrResource(Model):
        label = "Registrar IPs"
        model = RegistrarIPAddr
        fields = [Field(name="registrarid", input_=inputs.ForeignKey(model=Registrar)), "ipaddr"]

    class RegistrarZoneAccessResource(Model):
        label = "Registrar zones"
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

        fields = [Field(name="registrarid", input_=inputs.ForeignKey(model=Registrar)), Field(name="zone", input_=inputs.ForeignKey(model=Zone)), "fromdate"]

    label = "Registrators"
    icon = "fas fa-briefcase"
    resources = [RegistrarResource, RegistrarACLResource, RegistrarIPAddrResource, RegistrarZoneAccessResource]

@app.register
class Balance(Dropdown):
    class RegistrarBalanceResource(Model):
        label = "Registrar balance"
        model = RegistrarBalance
        fields = [Field(name="registrar", input_=inputs.ForeignKey(model=Registrar)), Field(name="zone", input_=inputs.ForeignKey(model=Zone)), "credit"]

        filters = [
            filters.Search(
                name="registrar__handle",
                label="Registrar",
                search_mode="contains",
                placeholder="Search for registrar",
            )
        ]

        async def get_toolbar_actions(self, request: Request) -> list[ToolbarAction]:
            return []

        async def get_actions(self, request: Request) -> list[Action]:
            return []

        async def get_bulk_actions(self, request: Request) -> list[Action]:
            return []

    label = "Balance"
    icon = "fas fa-dollar-sign"
    resources = [RegistrarBalanceResource]

@app.register
class Zones(Dropdown):
    class ZoneResource(Model):
        label = "Zone"
        model = Zone
        fields = ["fqdn"]

    class ZoneSOAResource(Model):
        label = "Zone SOA"
        model = ZoneSOA
        fields = [
            Field("zone", input_=inputs.ForeignKey(model=Zone)), 
            Field("ttl", label='TTL', display=displays.InputOnly(), input_=inputs.Text(placeholder="14400")),
            Field("serial", label="Zone serial", input_=inputs.Text(help_text="leave empty to use timestamp as serial")),
            Field("update_retr", display=displays.InputOnly()),
            Field("refresh", display=displays.InputOnly()),
            Field("expiry", display=displays.InputOnly()),
            Field("minimum", display=displays.InputOnly()),
            Field("hostmaster", input_=inputs.Text(placeholder="hostmaster.example.com")),
            Field("ns_fqdn", label='Nameserver', input_=inputs.Text(placeholder="ns1.example.com"))
        ]

        @classmethod
        def get_fields(self, is_display: bool = True):
            ret = super().get_fields(is_display=is_display)
            # override default behaviour, since it doesn't show primary keys
            ret[0] = Field(name="zone", input_=inputs.ForeignKey(model=Zone))
            return ret

    class ZoneDomainNameChecks(Model):
        label = "Domain name checks"
        model = DomainCheckMap
        fields = [Field(name="zone", input_=inputs.ForeignKey(model=Zone)), Field(name="checker", input_=inputs.ForeignKey(model=DomainChecks))]

    label = "Zones"
    icon = "fas fa-cog"
    resources = [ZoneResource, ZoneSOAResource, ZoneDomainNameChecks]
