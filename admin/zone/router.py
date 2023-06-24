from fastapi import APIRouter, HTTPException

import zone.schema as schema
import zone.zone as zone_funcs

router = APIRouter()

@router.get("/")
async def health_check():
    return {"status": "ok"}

@router.post("/zones", response_model=schema.Zone)
async def create_zone(post: schema.ZoneCreate):
    ret = await zone_funcs.create_zone(post)
    return ret

@router.get("/zones", response_model=list[schema.Zone])
async def get_zones(zone : str = None):
    zones = await zone_funcs.get_zones(zone)
    return zones

@router.get("/zones/{zone_id}", response_model=schema.Zone)
async def get_zone(zone_id : int): 
    zone_obj = await zone_funcs.get_zone(zone_id)
    if not zone_obj:
        raise HTTPException(status_code=404, detail="Zone not found")
    return zone_obj

@router.get("/zones/{zone_id}/pricelist", response_model=list[schema.ZonePriceList])
async def get_zone_pricelist(zone_id : int): 
    return await zone_funcs.get_zone_pricelist(zone_id)

@router.post("/zones/{zone_id}/pricelist", response_model=schema.ZonePriceList)
async def add_zone_pricelist(zone_id : int, post : schema.ZonePriceList): 
    return await zone_funcs.add_zone_pricelist(zone_id, post)
