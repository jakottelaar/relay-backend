# app/core/database.py
from contextlib import asynccontextmanager
from typing import AsyncGenerator, Annotated
from fastapi import Depends
from sqlmodel import SQLModel
from sqlmodel.ext.asyncio.session import AsyncSession
from sqlalchemy.ext.asyncio import create_async_engine
from sqlalchemy.pool import NullPool

from app.core.config import settings

# Create async engine
engine = create_async_engine(
    settings.POSTGRES_URI,
    echo=settings.DEBUG,
    future=True,
    poolclass=NullPool,
)


async def get_async_session() -> AsyncGenerator[AsyncSession, None]:
    async_session = AsyncSession(engine, expire_on_commit=False)
    try:
        yield async_session
    finally:
        await async_session.close()


# Type for dependency injection
AsyncSessionDep = Annotated[AsyncSession, Depends(get_async_session)]
