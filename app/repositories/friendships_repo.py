from sqlmodel import Session, select
from app.core.exceptions import AppException, ResourceNotFoundException
from app.models.friendships import Friendship
from app.schemas.friendships import FriendshipCreate
from sqlalchemy.exc import SQLAlchemyError


class FriendshipsRepository:
    def __init__(self, session: Session):
        self.session = session

    def create(self, friendship: FriendshipCreate) -> Friendship:
        try:
            db_friendship = Friendship(**friendship.model_dump())
            self.session.add(db_friendship)
            self.session.commit()
            self.session.refresh(db_friendship)
            return db_friendship
        except SQLAlchemyError as e:
            self.session.rollback()
            raise AppException(code=500, message="Database error occurred") from e

    def get_friendship(self, user_id: str, friend_id: str) -> Friendship:
        statement = select(Friendship).where(
            (Friendship.user_id == user_id) & (Friendship.friend_id == friend_id)
        )
        return self.session.exec(statement).first()

    def get_friendship_by_id(self, id: str) -> Friendship:
        friendship = self.session.exec(
            select(Friendship).where(Friendship.id == id)
        ).first()
        if not friendship:
            raise ResourceNotFoundException("Friendship", id)
        return friendship

    def delete_friendship(self, friendship: Friendship) -> None:
        try:
            self.session.delete(friendship)
            self.session.commit()
        except SQLAlchemyError as e:
            self.session.rollback()
            raise AppException(code=500, message="Database error occurred") from e
