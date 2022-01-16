from __future__ import annotations

from pol.api.v0.depends.page.base import BasePage


class OffsetPage(BasePage):
    page: int
    size: int
    limit: int
    # todo: add ordering (key, direction)

    @property
    def offset(self) -> int:
        return self.page * self.size

    @classmethod
    def from_request(
        cls,
        q: str | None = None,
        page: int | None = 0,
        size: int | None = 0,
        limit: int | None = 0,
    ) -> OffsetPage:
        return OffsetPage(page=page, size=size, limit=limit)

    def modify_query(self, query) -> None:
        query.limit(self.limit).offset(self.offset)
