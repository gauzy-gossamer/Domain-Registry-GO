import re

from datetime import date
from typing import Optional, Union
from ipaddress import IPv4Address, IPv6Address
from pydantic import BaseModel, validator

def check_fqdn(value: str) -> str:
    value = value.lower().strip('.')
    if not re.match(r'^[\w\-\.]{2,255}$', value):
        raise ValueError("incorrect zone fqdn")
    return value

class ZoneBase(BaseModel):
    fqdn: str
    ex_period_min: int
    ex_period_max: int

    @validator("fqdn")
    def check_fqdn(cls, value: str) -> str:
        return check_fqdn(value)

class Zone(ZoneBase):
    """ Return response data """
    id: int

class ZoneCreate(ZoneBase):
    """ create zone object """
    pass

class ZoneSoa(BaseModel):
    zone: Optional[int]
    serial: Optional[int]
    ttl: int
    refresh: int
    update_retr: int
    expiry: int
    minimum: int
    hostmaster: str
    ns_fqdn: str

    @validator("hostmaster", "ns_fqdn")
    def check_fqdn(cls, value: str) -> str:
        return check_fqdn(value)

class ZoneNs(BaseModel):
    id: Optional[int]
    zone: Optional[int]
    fqdn: str
    addrs: list[Union[IPv4Address, IPv6Address]]

    @validator("fqdn")
    def check_fqdn(cls, value: str) -> str:
        return check_fqdn(value)

class ZonePriceList(BaseModel):
    """ zone price """
    zoneid: Optional[int]
    fqdn: Optional[str]
    operation: str
    valid_from: date
    price: float

class DomainChecker(BaseModel):
    """ domain checker """
    id: Optional[int]
    name: str
    description: Optional[str]
