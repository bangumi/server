from typing import Any, Dict, List, Optional

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
    role: PersonRole
    summary: str
    subjects: List[int]
    locked: bool

    wiki: Optional[Dict[str, Any]] = Field(
        None,
        description="server parsed infobox, a map from key to string or tuple\n"
        "null if server infobox is not valid",
    )

    img: Optional[str] = None

    gender: Optional[int] = Field(None, description="parsed from wiki, maybe null")
    blood_type: Optional[int] = Field(None, description="parsed from wiki, maybe null")
    birth_year: Optional[int] = Field(None, description="parsed from wiki, maybe null")
    birth_mon: Optional[int] = Field(None, description="parsed from wiki, maybe null")
    birth_day: Optional[int] = Field(None, description="parsed from wiki, maybe null")
