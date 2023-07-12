import logging
from environs import Env 
from datetime import datetime, timedelta
from fastapi_mail import ConnectionConfig, FastMail, MessageSchema, MessageType
from fastapi_mail.errors import ConnectionErrors
from models.database import database

from mailer.models import mail_request_table, mail_request_type_table

import mailer.mail_requests as mail_requests

env = Env()
env.read_env()
logger = logging.getLogger('uvicorn')

conf = ConnectionConfig(
    MAIL_USERNAME = env("MAIL_USERNAME", ""),
    MAIL_PASSWORD = env("MAIL_PASSWORD", ""),
    MAIL_FROM = env("MAIL_FROM", "test@email.com"),
    MAIL_PORT = env.int("MAIL_PORT", 465),
    MAIL_SERVER = env("MAIL_SERVER", ""),
    MAIL_STARTTLS = env.bool("MAIL_STARTTLS", False),
    MAIL_SSL_TLS=env.bool("MAIL_SSL_TLS", True),
    USE_CREDENTIALS = env.bool("USE_CREDENTIALS", True),
    VALIDATE_CERTS = env.bool("VALIDATE_CERTS", False),
    TIMEOUT = 3,
)

async def get_pending_mails():
    query = mail_request_table.select().with_only_columns(
        mail_request_table.c.id, mail_request_table.c.object_id, mail_request_table.c.domain_id, mail_request_table.c.requested,
        mail_request_type_table.c.request_type, mail_request_type_table.c.subject, mail_request_type_table.c.template,
        mail_request_table.c.tries,
    ).join(mail_request_type_table, mail_request_table.c.request_type_id == mail_request_type_table.c.id
    ).where(
        mail_request_table.c.sent_mail == False, 
        mail_request_table.c.requested.between(datetime.now() - timedelta(days=3), datetime.now()),
        mail_request_table.c.tries <= 3, 
        mail_request_type_table.c.active == True
    )
    return await database.fetch_all(query)

async def update_mail_request(mr_id: int, sent_mail: bool, tries: int, mail_error: str = None) -> None:
    query = mail_request_table.update().values(
        sent_mail = sent_mail,
        tries = tries,
        mail_error = mail_error,
    ).where(mail_request_table.c.id == mr_id)
    await database.execute(query)

async def send_emails() -> None:
    if conf.MAIL_USERNAME == "":
        logger.warning("mailer is not configured")
        return

    fm = FastMail(conf)

    try:
        pending_emails = await get_pending_mails()
    except Exception as exc:
        logger.error(f'failed get_pending_mails: {exc}')
        return

    logger.info(f'mailer: processing {len(pending_emails)} pending mails')

    for email in pending_emails:
        try:
            if email.request_type not in mail_requests.mail_requests:
                logger.warning(f'mail request {email.request_type} not set up')
                continue

            mr = mail_requests.mail_requests[email.request_type](email)
            message = MessageSchema(
                subject=await mr.get_email_subject(),
                recipients=await mr.get_recipients(),
                body=await mr.get_email_body(),
                subtype=MessageType.plain)

            logger.info(f'sending {message}')
            await fm.send_message(message)
            await update_mail_request(email.id, True, email.tries)

        except ConnectionErrors as exc:
            logger.error(f'connection error: {exc}')
            await update_mail_request(email.id, False, email.tries + 1, mail_error = str(exc))
        except mail_requests.DomainNotFoundException as exc:
            logger.warning(f'domain not found: {exc}')
        except Exception as exc:
            logger.error(f'failed format message: {exc}')

