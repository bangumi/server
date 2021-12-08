from typing import Any, Dict, Optional

from pydantic import Field, BaseModel


class PersonRole(BaseModel):
    producer: bool
    mangaka: bool
    artist: bool
    seiyu: bool
    writer: bool
    illustrator: bool
    actor: bool


class Person(BaseModel):
    id: int
    name: str
    type: int
    infobox: str
    wiki: Optional[Dict[str, Any]] = Field(
        None,
        description="server parsed infobox, a map from key to string or tuple",
    )
    role: PersonRole
    summary: str
    img: Optional[str]
