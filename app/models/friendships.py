from sqlmodel import Field, SQLModel
import uuid
from datetime import datetime, timezone


class Friendship(SQLModel, table=True):

    __tablename__ = "friendships"

    id: uuid.UUID = Field(default=uuid.uuid4(), primary_key=True)
    user_id: str
    friend_id: str
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    updated_at: datetime = Field(
        default_factory=lambda: datetime.now(timezone.utc),
        sa_column_kwargs={"onupdate": lambda: datetime.now(timezone.utc)},
    )
