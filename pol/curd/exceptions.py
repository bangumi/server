from typing import Any

__all__ = ["NotFoundError"]


class NotFoundError(Exception):
    def __init__(self, details: Any = None):
        self.details = details
