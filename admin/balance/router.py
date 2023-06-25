from fastapi import APIRouter, HTTPException

import balance.schema as schema
import balance.balance as balance_funcs

router = APIRouter()

@router.get("/balance", response_model=list[schema.RegistrarCredit])
async def list_balance():
    return await balance_funcs.get_registrar_balance()

@router.get("/balance/{regid}", response_model=list[schema.RegistrarCredit])
async def get_balance(regid: int):
    return await balance_funcs.get_registrar_balance(regid)

@router.post("/balance/{regid}", response_model=list[schema.RegistrarCredit])
async def add_balance(regid: int, post: schema.AddRegistrarCredit):
    return await balance_funcs.add_registrar_balance(regid, post)

@router.get("/invoice", response_model=list[schema.RegistrarInvoice])
async def list_invoice():
    return await balance_funcs.get_registrar_invoice()
