from datetime import date
from typing import Optional
from pydantic import BaseModel

class Zone(BaseModel):
    """ Return response data """
    id: int
    fqdn: str
    ex_period_min: int
    ex_period_max: int

class ZoneCreate(BaseModel):
    """ create zone object """
    fqdn: str
    ex_period_min: int
    ex_period_max: int

class ZonePriceList(BaseModel):
    """ zone price """
    zoneid: Optional[int]
    fqdn: Optional[str]
    operation: str
    valid_from: date
    price: float
