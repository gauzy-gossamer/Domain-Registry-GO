from fastapi import Depends
from starlette.requests import Request

from fastapi_admin.app import app
from fastapi_admin.depends import get_resources
from fastapi_admin.template import templates
from fastapi_admin.routes.resources import create_view, create
from panel.resources import RegistrarTransactionResource

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

# there are no pages for RegistrarTransactionResource
# we add balance with these actions
@app.get("/registrarbalance/add_balance")
async def create_balance(
    request: Request,
    resources=Depends(get_resources),
):
    resource = 'registrarbalance'
    model_resource = RegistrarTransactionResource
    return await create_view(**locals())

@app.post("/registrarbalance/add_balance")
async def create_balance(
    request: Request,
    resources=Depends(get_resources),
):
    resource = 'registrarbalance'
    model_resource = RegistrarTransactionResource
    model = model_resource.model
    return await create(**locals())
