import uvicorn
from models.database import database
from fastapi import FastAPI, status
from fastapi.encoders import jsonable_encoder
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from fastapi_utils.tasks import repeat_every

from zone.router import router as zone_router
from registrar.router import router as registrar_router
from balance.router import router as balance_router
import mailer.mailer as mailer

from panel.main import create_admin_app


app = FastAPI()
create_admin_app(app)


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

@app.on_event("startup")
@repeat_every(seconds=60*2)
async def send_emails():
    await mailer.send_emails()

app.include_router(zone_router)
app.include_router(registrar_router)
app.include_router(balance_router)

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8088)
