from app.core.database import Base
from sqlalchemy import Column, UUID, DateTime, String


class Friendship(Base):
    __tablename__ = "friendships"

    id = Column(UUID, primary_key=True)
    user_id = Column(String, index=True, nullable=False)
    friend_id = Column(String, index=True, nullable=False)
    created_at = Column(DateTime, nullable=False, server_default=("now()"))
    updated_at = Column(DateTime, nullable=False, server_default=("now()"))
