from typing import Any, Optional
from fastapi_admin.resources import Model, ComputeField
from fastapi_admin.widgets import inputs
from panel.models import Zone

# multiple foreign key values proveded by related_fields list
class ForeignKeyRelated(inputs.ForeignKey):
    def __init__(
            self,
            model: type[Model],
            related_fields: list[str],
            null: bool = False,
            disabled: bool = False,
    ):  
        super().__init__(null=null, disabled=disabled, model=model)
        self.model = model
        self.related_fields = related_fields
        self.pk = self.model._meta.db_pk_column

    async def get_options(self):
        ret = await self.get_queryset()
        options = [(', '.join([x[k] for k in self.related_fields]), x[self.pk]) for x in ret]
        if self.context.get("null"):
            options = [("", "")] + options
        return options

    async def get_queryset(self):
        fields = self.related_fields + [self.pk]
        return await self.model.all().prefetch_related().values(*fields)


class IPAddress(inputs.Text):
    template = "widgets/inputs/ipaddress.html"


class Hostname(inputs.Text):
    template = "widgets/inputs/hostname.html"


# use pattern & title input properties
class Regex(inputs.Input):
    def __init__(
            self,
            regexp: str,
            error: str = '',
            help_text: Optional[str] = None,
            default: Any = None,
            null: bool = False,
            placeholder: str = "",
            disabled: bool = False,
    ):
        super().__init__(
            null=null,
            default=default,
            placeholder=placeholder,
            disabled=disabled,
            help_text=help_text,
            pattern=regexp,
            title=error,
        )
    template = "widgets/inputs/regex.html"


class Num(Regex):
    def __init__(
            self,
            help_text: Optional[str] = None,
            default: Any = None,
            null: bool = False,
            placeholder: str = "",
            disabled: bool = False,
    ):
        super().__init__(
            r"^\d+$",
            error="Should be a number.",
            null=null,
            default=default,
            placeholder=placeholder,
            disabled=disabled,
            help_text=help_text,
        )


class GetZoneName(ComputeField):
    async def get_value(self, request: 'Request', obj: dict):
        return await Zone.filter(id=obj.get('id')).first()
