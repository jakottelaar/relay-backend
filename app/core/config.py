from pydantic import field_validator
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    PROJECT_NAME: str
    CLERK_PEM_PUBLIC_KEY: str
    CLERK_API_KEY: str
    POSTGRES_URI: str
    # REDIS_URI: str

    @field_validator("POSTGRES_URI")
    def validate_postgres_uri(cls, v):
        if not v.startswith("postgresql://"):
            raise ValueError("Invalid postgres uri")
        return v

    # @field_validator("REDIS_URI")
    # def validate_redis_uri(cls, v):
    #     if not v.startswith("redis://"):
    #         raise ValueError("Invalid redis uri")
    #     return v
    class Config:
        env_file = ".env.local"


settings = Settings()
