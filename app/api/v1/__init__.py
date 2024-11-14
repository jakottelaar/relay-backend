from fastapi import APIRouter

router = APIRouter(prefix="/v1")

from app.api.v1.users import router as users_router

router.include_router(users_router)
