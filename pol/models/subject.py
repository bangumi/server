from typing import Optional

from pydantic import BaseModel


class Subject(BaseModel):
    id: int
    type: int
    name: str
    name_cn: str
    summary: str
    nsfw: bool
    date: Optional[str]  # air date in `YYYY-MM-DD` format"
    platform: int  # TV, Web, 欧美剧, PS4...
    image: str
    infobox: str

    redirect: int

    ban: int

    @property
    def banned(self) -> bool:
        """redirected/merged subject"""
        return self.ban == 1

    @property
    def locked(self) -> bool:
        return self.ban == 2
