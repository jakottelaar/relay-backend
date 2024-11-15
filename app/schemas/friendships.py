from pydantic import BaseModel
from datetime import datetime
from uuid import UUID


class FriendshipCreate(BaseModel):
    user_id: str
    friend_id: str


class FriendshipResponse(BaseModel):
    id: UUID
    user_id: str
    friend_id: str
    created_at: datetime
    updated_at: datetime
