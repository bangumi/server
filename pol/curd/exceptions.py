from typing import Any


class NotFoundError(Exception):
    def __init__(self, details: Any = None):
        self.details = details
