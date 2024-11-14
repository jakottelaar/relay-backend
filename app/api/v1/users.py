from typing import List
from fastapi import APIRouter, HTTPException
from app.services.clerk_service import ClerkService
from app.schemas.users import UserOut

router = APIRouter(prefix="/users", tags=["users"])

clerk_service = ClerkService()


@router.get("/search", response_model=List[UserOut])
def search_users():
    try:
        # Retrieve list of users as dictionaries from the Clerk service
        users = clerk_service.get_users_by_username(usernames=["testUserName"])

        # Convert each dictionary to a UserOut model instance
        user_out_list = [UserOut(**user) for user in users]

        return user_out_list
    except Exception as e:
        raise HTTPException(
            status_code=500, detail="An error occurred while fetching users."
        )
