from fastapi import APIRouter

router = APIRouter(prefix="/v1")

from app.api.v1.users import router as users_router
from app.api.v1.friendships import router as friendships_router

router.include_router(users_router)
router.include_router(friendships_router)
