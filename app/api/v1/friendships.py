from uuid import UUID
from fastapi import APIRouter, HTTPException, Depends
from sqlmodel import Session
from app.api.payload import ErrorResponse
from app.core.error_handler import handle_app_exception
from app.core.exceptions import AppException
from app.schemas.friendships import FriendshipCreate, FriendshipResponse
from app.services.friendships_service import FriendshipService
from app.core.database import get_session


router = APIRouter(prefix="/friendships", tags=["friendships"])


@router.post(
    "/",
    response_model=FriendshipResponse,
    status_code=201,
    responses={
        422: {"model": ErrorResponse},
        400: {"model": ErrorResponse},
        409: {"model": ErrorResponse},
        500: {"model": ErrorResponse},
    },
)
async def create_friendship(
    friendship_req: FriendshipCreate, session: Session = Depends(get_session)
):
    try:
        friendship_service = FriendshipService(session)
        return friendship_service.create_friendship(friendship_req)
    except AppException as e:
        raise HTTPException(
            status_code=e.code, detail=handle_app_exception(e).model_dump()
        )


@router.get(
    "/{id}",
    response_model=FriendshipResponse,
    status_code=200,
    responses={404: {"model": ErrorResponse}, 500: {"model": ErrorResponse}},
)
def get_friendship(id: UUID, session: Session = Depends(get_session)):
    try:
        friendship_service = FriendshipService(session)
        return friendship_service.get_friendship_by_id(id)
    except AppException as e:
        raise HTTPException(
            status_code=e.code, detail=handle_app_exception(e).model_dump()
        )


@router.delete(
    "/{id}",
    status_code=200,
    responses={404: {"model": ErrorResponse}, 500: {"model": ErrorResponse}},
)
def delete_friendship(id: UUID, session: Session = Depends(get_session)):
    try:
        friendship_service = FriendshipService(session)
        friendship_service.delete_friendship(id)
        return {"message": "Friendship deleted successfully."}
    except AppException as e:
        raise HTTPException(
            status_code=e.code, detail=handle_app_exception(e).model_dump()
        )
