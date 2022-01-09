from typing import TypeVar

from pol.db.tables import Base
from . import ep
from .base import count

T = TypeVar("T", bound=Base)

__all__ = ["ep", "count"]
