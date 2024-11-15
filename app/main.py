from fastapi import FastAPI
from fastapi.exceptions import RequestValidationError
from app.api import router as api_router
from app.core.error_handler import validation_exception_handler
from app.middleware.clerk_auth import VerifyTokenMiddleware

app = FastAPI()

app.include_router(api_router)

# app.add_middleware(VerifyTokenMiddleware)

app.add_exception_handler(RequestValidationError, validation_exception_handler)


@app.get("/")
def read_root():
    return {"Hello": "World"}
