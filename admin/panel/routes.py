from fastapi import Depends
from starlette.requests import Request

from fastapi_admin.app import app
from fastapi_admin.depends import get_resources
from fastapi_admin.template import templates
from tortoise.functions import Sum
from panel.resources import RegistrarTransactionResource
from panel.models import RegistrarBalance
from tortoise.transactions import in_transaction
from fastapi_admin.responses import redirect

@app.get("/")
async def home(
    request: Request,
    resources=Depends(get_resources),
):
    qs = await RegistrarBalance.all().annotate(total=Sum("credit")).group_by("registrar__id", "registrar__handle").values("registrar__id", "registrar__handle", "total")

    return templates.TemplateResponse(
        "dashboard.html",
        context={
            "request": request,
            "resources": resources,
            "labels":["Registrar", "", "", "", "Current balance"],
            "values":qs,
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
    inputs = await model_resource.get_inputs(request)
    context = {
        "request": request,
        "resources": resources,
        "resource_label": model_resource.label,
        "resource": resource,
        "inputs": inputs,
        "model_resource": model_resource,
        "page_title": model_resource.page_title,
        "page_pre_title": model_resource.page_pre_title,
    }
    return templates.TemplateResponse(
        "add_balance.html",
        context=context,
    )

@app.post("/registrarbalance/add_balance")
async def create_balance(
    request: Request,
    resources=Depends(get_resources),
):
    resource = 'registrarbalance'
    model_resource = RegistrarTransactionResource
    model = model_resource.model
    inputs = await model_resource.get_inputs(request)
    form = await request.form()
    data, _ = await model_resource.resolve_data(request, form)
    async with in_transaction() as conn:
        obj = await model.create(**data, using_db=conn)
    if "save" in form.keys():
        return redirect(request, "list_view", resource=resource)
    context = {
        "request": request,
        "resources": resources,
        "resource_label": model_resource.label,
        "resource": resource,
        "inputs": inputs,
        "model_resource": model_resource,
        "page_title": model_resource.page_title,
        "page_pre_title": model_resource.page_pre_title,
    }
    return templates.TemplateResponse(
        "add_balance.html",
        context=context,
    )
