from fastapi import HTTPException, status, Request
from fastapi.responses import JSONResponse
from fastapi.security import HTTPBearer
import jwt
from starlette.middleware.base import BaseHTTPMiddleware
from app.core.config import settings


security = HTTPBearer()


class VerifyTokenMiddleware(BaseHTTPMiddleware):
    """
    Middleware to verify JWT token from Clerk
    """

    async def dispatch(self, request: Request, call_next):
        try:
            # Get the Authorization header from the request
            auth_header = request.headers.get("Authorization")
            if not auth_header:
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="No authorization header found",
                )

            # Ensure the authorization scheme is 'Bearer'
            scheme, token = auth_header.split()
            if scheme.lower() != "bearer":
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Invalid authentication scheme",
                )

            # Decode the token using Clerk's public key
            decoded_token = jwt.decode(
                token, key=settings.CLERK_PEM_PUBLIC_KEY, algorithms=["RS256"]
            )

            # Add the decoded token and user_id to the request state
            request.state.user_id = decoded_token.get("sub")

            # Proceed to the next middleware or route handler
            response = await call_next(request)
            return response

        except jwt.InvalidTokenError as e:
            return JSONResponse(
                status_code=status.HTTP_401_UNAUTHORIZED,
                content={"detail": f"Invalid token: {str(e)}"},
            )
        except Exception as e:
            return JSONResponse(
                status_code=status.HTTP_401_UNAUTHORIZED,
                content={"detail": f"Authentication failed: {str(e)}"},
            )


# Dependency to get current user ID
async def get_current_user_id(request: Request) -> str:
    """
    Dependency that returns the current user ID from the request state
    """
    if not hasattr(request.state, "user_id"):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="Not authenticated"
        )
    return request.state.user_id
