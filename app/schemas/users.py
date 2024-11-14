from pydantic import BaseModel, EmailStr


class UserBase(BaseModel):
    id: str
    email: EmailStr
    username: str
    profile_image_url: str


class UserOut(UserBase):
    pass
