from clerk_backend_api import Clerk, models
from typing import List, Optional
from app.core.config import settings


class ClerkService:
    def __init__(self):
        self.client = Clerk(bearer_auth=settings.CLERK_API_KEY)

    def get_users_by_username(self, usernames: Optional[List[str]]):
        try:
            # Retrieve user data from Clerk API
            res = self.client.users.list(username=usernames)

            # Filter and structure the user data
            users = [
                {
                    "id": user.id,
                    "email": user.email_addresses[0].email_address,
                    "username": user.username,
                    "profile_image_url": user.profile_image_url,
                }
                for user in res
            ]
            return users
        except models.ClerkErrors as e:
            raise e
        except models.SDKError as e:
            raise e
