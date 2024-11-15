from typing import List

from fastapi import Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from pydantic import BaseModel
from app.api.payload import ErrorDetail, ErrorResponse
from app.core.exceptions import AppException


class ValidationError(BaseModel):
    field: str
    message: str


def handle_app_exception(exc: AppException) -> ErrorResponse:
    return ErrorResponse(
        error=ErrorDetail(code=exc.code, message=exc.message, errors=exc.errors)
    )


def format_validation_errors(errors: List[dict]) -> List[ValidationError]:
    """Convert Pydantic error format to our API format"""
    formatted_errors = []
    for error in errors:
        # Get the field name from the error location
        field = " -> ".join(str(loc) for loc in error["loc"] if loc != "body")
        formatted_errors.append(ValidationError(field=field, message=error["msg"]))
    return formatted_errors


async def validation_exception_handler(
    request: Request, exc: RequestValidationError
) -> JSONResponse:
    """Handle Pydantic validation errors and return consistent error format"""
    formatted_errors = format_validation_errors(exc.errors())

    error_dict = {}
    for error in formatted_errors:
        if error.field not in error_dict:
            error_dict[error.field] = []
        error_dict[error.field].append(error.message)

    return JSONResponse(
        status_code=422,
        content=ErrorResponse(
            error=ErrorDetail(code=422, message="Validation error", errors=error_dict)
        ).model_dump(),
    )
