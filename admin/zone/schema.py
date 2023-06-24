import re

from datetime import date
from typing import Optional
from pydantic import BaseModel, validator

class ZoneBase(BaseModel):
    fqdn: str
    ex_period_min: int
    ex_period_max: int

    @validator("fqdn")
    def check_fqdn(cls, value: str):
        value = value.lower().strip('.')
        if not re.match(r'^[\w\-\.]{2,255}$', value):
            raise ValueError("incorrect zone fqdn")

        return value

class Zone(ZoneBase):
    """ Return response data """
    id: int

class ZoneCreate(ZoneBase):
    """ create zone object """
    pass

class ZonePriceList(BaseModel):
    """ zone price """
    zoneid: Optional[int]
    fqdn: Optional[str]
    operation: str
    valid_from: date
    price: float
