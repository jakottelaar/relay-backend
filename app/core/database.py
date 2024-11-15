from sqlmodel import create_engine, Session
from app.core.config import settings

engine = create_engine(settings.POSTGRES_URI)


def get_session():
    with Session(engine) as session:
        yield session
