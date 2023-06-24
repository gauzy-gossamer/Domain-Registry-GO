import re
from datetime import date
from typing import Optional, Union
from ipaddress import IPv4Address, IPv6Address

from pydantic import BaseModel, EmailStr, HttpUrl, validator, Field

class RegistrarBase(BaseModel):
    handle: str
    name: Optional[str]
    intpostal: Optional[str]
    intaddress: Optional[list[str]]
    locpostal: Optional[str]
    locaddress: Optional[list[str]]
    legaladdress: Optional[list[str]]
    taxpayerNumbers: Optional[str]
    telephone: Optional[list[str]]
    fax: Optional[list[str]]
    email: Optional[list[EmailStr]]
    notify_email: Optional[list[EmailStr]]
    info_email: Optional[list[EmailStr]]
    url: Optional[HttpUrl]
    www: Optional[HttpUrl]
    system: Optional[bool]
    epp_requests_limit: Optional[int]

class Registrar(RegistrarBase):
    """ Registrar object """
    id: int
    object_id: int

class RegistrarCreate(RegistrarBase):
    """ Create registrar """
    pass

class RegistrarAclBase(BaseModel):
    cert: str
    password: str

    @validator("cert")
    def check_cert(cls, value: str):
        value = value.upper()
        parts = value.split(':')
        if len(parts) != 16:
            raise ValueError("incorrect cert")
        if not all([re.match(r'^[0-9A-F]{2}$', v) for v in parts]):
            raise ValueError("incorrect cert")
        return value

class RegistrarAclCreate(RegistrarAclBase):
    pass

class RegistrarAcl(RegistrarAclBase):
    id: int
    registrarid: int

class RegistrarIpAddrBase(BaseModel):
    ipaddr: Union[IPv4Address, IPv6Address]

class RegistrarIpAddr(RegistrarIpAddrBase):
    id: int
    registrarid: int

class RegistrarZones(BaseModel):
    zone: str
    fromdate: Optional[date]