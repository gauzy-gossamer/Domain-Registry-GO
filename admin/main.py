import uvicorn
from models.database import database
from fastapi import FastAPI, status
from fastapi.encoders import jsonable_encoder
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse

from zone.router import router as zone_router
from registrar.router import router as registrar_router
from balance.router import router as balance_router

app = FastAPI()

@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request, exc):
    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content=jsonable_encoder({"detail": exc.errors(), "Error": "Validation error"}),
    )

@app.on_event("startup")
async def startup():
    await database.connect()


@app.on_event("shutdown")
async def shutdown():
    await database.disconnect()

app.include_router(zone_router)
app.include_router(registrar_router)
app.include_router(balance_router)

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8088)
