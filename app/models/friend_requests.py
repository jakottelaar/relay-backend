from sqlalchemy import UUID, Column, DateTime, String
from app.core.database import Base


class FriendRequests(Base):
    __tablename__ = "friend_requests"

    id = Column(UUID, primary_key=True)
    sender_id = Column(String, index=True, nullable=False)
    receiver_id = Column(String, index=True, nullable=False)
    created_at = Column(DateTime, nullable=False, server_default=("now()"))
    updated_at = Column(DateTime, nullable=False, server_default=("now()"))
