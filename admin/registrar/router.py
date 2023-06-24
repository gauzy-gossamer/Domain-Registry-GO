from fastapi import APIRouter, HTTPException

import registrar.schema as schema
import registrar.registrar as reg_funcs

router = APIRouter()

@router.get("/registrars", response_model=list[schema.Registrar])
async def get_registrars(handle: str = None):
    registrars = await reg_funcs.get_registrars(handle=handle)
    return registrars

@router.get("/registrars/{regid}", response_model=schema.Registrar)
async def get_registrar(regid : int):
    reg_obj = await reg_funcs.get_registrar(regid)
    if not reg_obj:
        raise HTTPException(status_code=404, detail="Registrar not found")
    return reg_obj

@router.post("/registrars", response_model=schema.Registrar)
async def create_registrar(post: schema.RegistrarCreate):
    reg = await reg_funcs.create_registrar(post)
    return reg

@router.get("/registrars/{regid}/acl", response_model=schema.RegistrarAcl)
async def get_registraracl(regid : int):
    regacl_obj = await reg_funcs.get_registraracl(regid)
    if not regacl_obj:
        raise HTTPException(status_code=404, detail="Registrar not found")
    return regacl_obj

@router.post("/registrars/{regid}/acl", response_model=schema.RegistrarAcl)
async def create_registraracl(regid : int, post : schema.RegistrarAclCreate):
    return await reg_funcs.create_registraracl(regid, post)

@router.put("/registrars/{regid}/acl", response_model=schema.RegistrarAcl)
async def update_registraracl(regid : int, post : schema.RegistrarAclCreate):
    return await reg_funcs.update_registraracl(regid, post)

@router.get("/registrars/{regid}/ips", response_model=list[schema.RegistrarIpAddr])
async def get_registrar_ips(regid : int):
    return await reg_funcs.get_registrar_ips(regid)

@router.put("/registrars/{regid}/ips", response_model=list[schema.RegistrarIpAddrBase])
async def update_registrar_ips(regid : int, post : list[schema.RegistrarIpAddrBase]):
    regips_obj = await reg_funcs.update_registrar_ips(regid, post)
    return regips_obj

@router.get("/registrars/{regid}/zones", response_model=list[schema.RegistrarZones])
async def get_registrar_ips(regid : int):
    return await reg_funcs.get_registrar_zones(regid)

@router.put("/registrars/{regid}/zones", response_model=list[schema.RegistrarZones])
async def update_registrar_ips(regid : int, post : list[schema.RegistrarZones]):
    return await reg_funcs.add_registrar_zones(regid, post)

@router.delete("/registrars/{regid}/zones", response_model=list[schema.RegistrarZones])
async def update_registrar_ips(regid : int, post : list[schema.RegistrarZones]):
    return await reg_funcs.del_registrar_zones(regid, post)