import uvicorn
from models.database import database
from zone.router import router as zone_router
from registrar.router import router as registrar_router
from fastapi import FastAPI

app = FastAPI()


@app.on_event("startup")
async def startup():
    await database.connect()


@app.on_event("shutdown")
async def shutdown():
    await database.disconnect()


app.include_router(zone_router)
app.include_router(registrar_router)

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8088)
