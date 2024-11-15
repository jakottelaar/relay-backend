from fastapi import HTTPException
from app.core.exceptions import ResourceAlreadyExistsException, ValidationException
from app.repositories.friendships_repo import FriendshipsRepository
from app.schemas.friendships import FriendshipCreate, FriendshipResponse
from sqlmodel import Session


class FriendshipService:
    def __init__(self, session: Session):
        self.repository = FriendshipsRepository(session)

    def create_friendship(self, friendship: FriendshipCreate) -> FriendshipResponse:
        if friendship.user_id == friendship.friend_id:
            raise ValidationException(message="Cannot be friends with yourself")

        existing = self.repository.get_friendship(
            friendship.user_id, friendship.friend_id
        )
        if existing:
            raise ResourceAlreadyExistsException(message="Friendship already exists")

        db_friendship = self.repository.create(friendship)

        return FriendshipResponse(
            id=db_friendship.id,
            user_id=db_friendship.user_id,
            friend_id=db_friendship.friend_id,
            created_at=db_friendship.created_at,
            updated_at=db_friendship.updated_at,
        )

    def delete_friendship(self, id: str):
        friendship = self.repository.get_friendship_by_id(id)
        self.repository.delete_friendship(friendship)
