from pydantic import BaseModel
from typing import Optional, Dict, Any


class Response(BaseModel):
    content: dict


class ErrorDetail(BaseModel):
    code: int
    message: str
    errors: Optional[Dict[str, Any]] = None


class ErrorResponse(BaseModel):
    error: ErrorDetail
