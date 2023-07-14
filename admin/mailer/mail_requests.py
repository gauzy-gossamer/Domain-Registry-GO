import json
from abc import ABC, abstractmethod
from jinja2 import Environment, BaseLoader
from models.database import database

from registrar.models import registrars_table
from registrar.registrar import get_registrar
from balance.balance import get_registrar_balance

class DomainNotFoundException(Exception):
    pass

class MailRequest(ABC):
    def __init__(self, email):
        self.email = email

    async def get_email_subject(self) -> str:
        rtemplate = Environment(loader=BaseLoader).from_string(self.email.subject)
        return rtemplate.render(await self._get_params())

    async def get_email_body(self) -> str:
        rtemplate = Environment(loader=BaseLoader).from_string(self.email.template)
        return rtemplate.render(await self._get_params())

    @abstractmethod
    async def get_recipients(self) -> list:
        ...

    @abstractmethod
    async def _get_params(self) -> dict:
        ...

class MailLowCredit(MailRequest):
    async def get_recipients(self) -> list:
        registrar = await get_registrar(self.email.object_id)
        emails = json.loads(registrar.email)
        if len(emails) == 0:
            raise Exception("registrar {self.email.object_id} mails are not setup")
        return emails

    async def _get_params(self) -> dict:
        balance = await get_registrar_balance(self.email.object_id)
        credit = sum([b.credit for b in balance])
        return {
            'credit':credit,
            'request_datetime':self.email.requested,
        }

class MailNewTransfer(MailRequest):
    async def get_recipients(self) -> list:
        registrar = await get_registrar(self.email.object_id)
        emails = json.loads(registrar.email)
        if len(emails) == 0:
            raise Exception("registrar {self.email.object_id} mails are not setup")
        return emails

    async def _get_params(self) -> dict:
        object_data = await database.fetch_one('''SELECT obr.name, obr.crdate, d.exdate::date as free_date, r.handle as registrar, acid.handle as acid, et.name as state
                      FROM object_registry obr JOIN object o on obr.id=o.id JOIN registrar r on o.clid=r.id JOIN domain d on obr.id=d.id 
                        JOIN epp_transfer_request rt on obr.id=rt.domain_id JOIN registrar acid on acid.id=rt.acquirer_id
                        JOIN enum_transfer_states et on rt.status = et.id
                      WHERE obr.id = :id;''',
                      values={'id':self.email.domain_id})

        if object_data is None:
            raise DomainNotFoundException(f"domain {self.email.domain_id} not found")

        return {
            **object_data,
            'request_datetime':self.email.requested,
        }

mail_requests = {
    "transfer_new":MailNewTransfer,
    "lowcredit":MailLowCredit,
}
