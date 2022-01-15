from typing import Optional, Protocol

from pol.models import Permission


class Role(Protocol):
    permission: Permission

    def allow_nsfw(self) -> bool:
        """if this user can see nsfw contents"""
        raise NotImplementedError()

    def get_user_id(self) -> Optional[int]:
        raise NotImplementedError()
