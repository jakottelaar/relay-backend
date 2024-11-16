# database.py
from sqlmodel import create_engine, Session
from sqlalchemy.exc import SQLAlchemyError
from sqlalchemy import text
import logging
import time
from app.core.config import settings

logger = logging.getLogger(__name__)


def wait_for_database(max_retries: int = 5, retry_delay: int = 5) -> bool:
    """
    Simple check to verify database connection before starting the application.

    Args:
        max_retries: Number of connection attempts before giving up (default: 5)
        retry_delay: Seconds to wait between retries (default: 5)
    """
    logger.info("Checking database connection...")

    for attempt in range(max_retries + 1):
        try:
            with Session(engine) as session:
                result = session.exec(text("SELECT 1"))
                result.first()

            logger.info("Successfully connected to database")
            return True

        except SQLAlchemyError as e:
            if attempt == max_retries:
                logger.error(
                    f"Failed to connect to database after {max_retries} attempts: {str(e)}"
                )
                return False

            logger.warning(
                f"Database connection attempt {attempt + 1} failed. Retrying in {retry_delay} seconds...\nError: {str(e)}"
            )
            time.sleep(retry_delay)

    return False


# Create the engine once at module level
engine = create_engine(
    settings.POSTGRES_URI,
    pool_pre_ping=True,
    pool_recycle=3600,  # Recycle connections after 1 hour
)


def get_session():
    with Session(engine) as session:
        yield session
