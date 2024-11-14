from clerk_backend_api import Clerk
from typing import List, Optional
from clerk_backend_api import Clerk
from dotenv import load_dotenv
import os


class ClerkService:
    def __init__(self):
        load_dotenv(".env.local")
        self.client = Clerk(bearer_auth=os.getenv("CLERK_API_KEY"))

    async def get_user_with_query(self, query: Optional[str]):
        users = self.client.users.list(query=query)
        return users

    async def get_users_by_email(self, emails: Optional[List[str]]):
        users = self.client.users.list(email_address=emails)
        return users
