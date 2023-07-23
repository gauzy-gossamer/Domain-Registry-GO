from fastapi import Depends
from starlette.requests import Request

from fastapi_admin.app import app
from fastapi_admin.depends import get_resources
from fastapi_admin.template import templates


@app.get("/")
async def home(
    request: Request,
    resources=Depends(get_resources),
):
    return templates.TemplateResponse(
        "dashboard.html",
        context={
            "request": request,
            "resources": resources,
            "resource_label": "Dashboard",
            "page_pre_title": "overview",
            "page_title": "Dashboard",
        },
    )
