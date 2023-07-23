import os
import redis.asyncio as redis
import uvicorn
from fastapi import FastAPI
from starlette.responses import RedirectResponse
from starlette.status import HTTP_401_UNAUTHORIZED, HTTP_403_FORBIDDEN, HTTP_404_NOT_FOUND, HTTP_500_INTERNAL_SERVER_ERROR
from tortoise.contrib.fastapi import register_tortoise
from fastapi_admin.app import app as admin_app
from fastapi_admin.exceptions import (
    forbidden_error_exception,
    not_found_error_exception,
    server_error_exception,
    unauthorized_error_exception,
)

from models.database import TORTOISE_DATABASE_URL, REDIS_URL
from panel.models import Admin
from panel.providers import LoginProvider

BASE_DIR = os.path.dirname(os.path.abspath(__file__))

def create_admin_app(app):
    admin_app.add_exception_handler(HTTP_500_INTERNAL_SERVER_ERROR, server_error_exception)
    admin_app.add_exception_handler(HTTP_404_NOT_FOUND, not_found_error_exception)
    admin_app.add_exception_handler(HTTP_403_FORBIDDEN, forbidden_error_exception)
    admin_app.add_exception_handler(HTTP_401_UNAUTHORIZED, unauthorized_error_exception)

    @app.on_event("startup")
    async def startup():
        r = redis.from_url(
            REDIS_URL,
            decode_responses=True,
            encoding="utf8",
        )
        await admin_app.configure(
            logo_url="https://preview.tabler.io/static/logo-white.svg",
            template_folders=[os.path.join(BASE_DIR, "templates")],
            favicon_url="https://raw.githubusercontent.com/fastapi-admin/fastapi-admin/dev/images/favicon.png",
            providers=[
                LoginProvider(
                    login_logo_url="https://preview.tabler.io/static/logo.svg",
                    admin_model=Admin,
                )
            ],
            redis=r,
        )

    app.mount("/admin", admin_app)
    register_tortoise(
        app,
        config={
            "connections": {"default": TORTOISE_DATABASE_URL},
            "apps": {
                "models": {
                    "models": ["panel.models"],
                    "default_connection": "default",
                }
            },
        },
        generate_schemas=False,
    )
