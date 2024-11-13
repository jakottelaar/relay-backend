from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from app.services.clerk_service import ClerkService

router = APIRouter(prefix="/users", tags=["users"])
clerk_service = ClerkService()

@router.get("/search")
async def get_user(
    query: Optional[str] = Query(None),
    emailAddresses: Optional[List[str]] = Query(None),
):
    if query:
        users = await clerk_service.get_user_with_query(query=query)
        return {"users": users}
    elif emailAddresses:
        users = await clerk_service.get_users_by_email(emails=emailAddresses)
        return {"users": users}
    else:
        raise HTTPException(status_code=400, detail="No query or email provided")