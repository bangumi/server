from pol.permission.types import UserPermState


class Role:
    """abstract class"""

    user_perm_state: UserPermState

    def __init__(self, user_perm_state: UserPermState = UserPermState()):
        self.user_perm_state = user_perm_state

    def allow_nsfw(self) -> bool:
        return self.user_perm_state.canViewNsfw


class GuestRole(Role):
    """this is a guest with only basic permission"""
