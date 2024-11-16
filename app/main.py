import sys
from fastapi import FastAPI
from fastapi.exceptions import RequestValidationError
from app.api import router as api_router
from app.core.database import wait_for_database
from app.core.error_handler import validation_exception_handler
from app.middleware.clerk_auth import VerifyTokenMiddleware
from app.core.config import settings
import logging

logger = logging.getLogger(__name__)


def start_application() -> FastAPI:

    if not wait_for_database():
        logger.critical("Could not connect to database. Exiting...")
        sys.exit(1)

    app = FastAPI()
    app.include_router(api_router)
    app.add_middleware(VerifyTokenMiddleware)
    app.add_exception_handler(RequestValidationError, validation_exception_handler)
    return app


app = start_application()
