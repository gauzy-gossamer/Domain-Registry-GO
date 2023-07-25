import uuid
from datetime import datetime, timedelta
from pytz import timezone
from starlette.requests import Request
from starlette.responses import RedirectResponse
from starlette.middleware.base import RequestResponseEndpoint
from starlette.status import HTTP_303_SEE_OTHER, HTTP_401_UNAUTHORIZED

from fastapi_admin import constants
from panel.models import AdminSession
from fastapi_admin.providers.login import UsernamePasswordProvider
from fastapi_admin.template import templates
from fastapi_admin.utils import check_password


# override default LoginProvider to drop dependency on redis
class LoginProvider(UsernamePasswordProvider):
    async def login(self, request: Request):
        form = await request.form()
        username = form.get("username")
        password = form.get("password")
        remember_me = form.get("remember_me")
        admin = await self.admin_model.get_or_none(username=username)
        if not admin or not check_password(password, admin.password):
            return templates.TemplateResponse(
                self.template,
                status_code=HTTP_401_UNAUTHORIZED,
                context={"request": request, "error": _("login_failed")},
            )
        response = RedirectResponse(url=request.app.admin_path, status_code=HTTP_303_SEE_OTHER)
        if remember_me == "on":
            expire = 3600 * 24 * 30
            response.set_cookie("remember_me", "on")
        else:
            expire = 3600
            response.delete_cookie("remember_me")
        token = uuid.uuid4().hex
        response.set_cookie(
            self.access_token,
            token,
            expires=expire,
            path=request.app.admin_path,
            httponly=True,
        )
        ex_datetime = datetime.now(timezone('UTC')) + timedelta(seconds=expire)
        await AdminSession.create(token=constants.LOGIN_USER.format(token=token), admin=admin, expire=ex_datetime)
        return response

    async def logout(self, request: Request):
        response = self.redirect_login(request)
        response.delete_cookie(self.access_token, path=request.app.admin_path)
        token = request.cookies.get(self.access_token)
        await AdminSession.filter(token=constants.LOGIN_USER.format(token=token)).delete()
        return response

    async def authenticate(
        self,
        request: Request,
        call_next: RequestResponseEndpoint,
    ):
        token = request.cookies.get(self.access_token)
        path = request.scope["path"]
        admin = None
        if token:
            token_key = constants.LOGIN_USER.format(token=token)
            admin_session = await AdminSession.get_or_none(token=token_key, expire__gt=datetime.now(timezone('UTC')))
            if admin_session is not None:
                admin = await self.admin_model.get_or_none(pk=admin_session.admin_id)
        request.state.admin = admin

        if path == self.login_path and admin:
            return RedirectResponse(url=request.app.admin_path, status_code=HTTP_303_SEE_OTHER)

        response = await call_next(request)
        return response
