from fastapi import HTTPException, status, Request
from fastapi.responses import JSONResponse
import jwt
from starlette.middleware.base import BaseHTTPMiddleware
from app.core.config import settings


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

            # Add the decoded token to the request state
            request.state.user = decoded_token

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
