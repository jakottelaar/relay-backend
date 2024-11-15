from typing import Optional, Dict, Any


class AppException(Exception):
    """Base exception for application"""

    def __init__(
        self, code: int, message: str, errors: Optional[Dict[str, Any]] = None
    ):
        self.code = code
        self.message = message
        self.errors = errors
        super().__init__(message)


class ResourceNotFoundException(AppException):
    def __init__(self, resource: str, identifier: Any):
        super().__init__(
            code=404, message=f"{resource} with identifier {identifier} not found"
        )


class ResourceAlreadyExistsException(AppException):
    def __init__(self, message: str):
        super().__init__(code=409, message=message)


class ValidationException(AppException):
    def __init__(self, message: str, errors: Optional[Dict[str, Any]] = None):
        super().__init__(code=400, message=message, errors=errors)
