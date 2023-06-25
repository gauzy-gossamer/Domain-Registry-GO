import datetime
from typing import Optional

from pydantic import BaseModel

class RegistrarCredit(BaseModel):
    handle: Optional[str]
    credit: float
    zone: str

class RegistrarCreditTransaction(BaseModel):
    id: int
    balance_change: float
    credit_id: int

class AddRegistrarCredit(BaseModel):
    balance_change: float
    zone: str

class RegistrarInvoice(BaseModel):
    handle: Optional[str]
    crdate: datetime.datetime
    balance: float
    zone: str
    comment: str
