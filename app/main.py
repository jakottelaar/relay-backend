from fastapi import FastAPI
from app.api import router as api_router
from app.middleware.clerk_auth import VerifyTokenMiddleware

app = FastAPI()

app.include_router(api_router)

app.add_middleware(VerifyTokenMiddleware)


@app.get("/")
def read_root():
    return {"Hello": "World"}
