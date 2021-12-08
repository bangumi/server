from typing import Any, Dict, TypeVar, Optional

from pydantic import Field, BaseModel

T = TypeVar("T", bound="PersonRole")


class PersonRole(BaseModel):
    producer: bool = Field(..., alias="prsn_producer")
    mangaka: bool = Field(..., alias="prsn_mangaka")
    artist: bool = Field(..., alias="prsn_artist")
    seiyu: bool = Field(..., alias="prsn_seiyu")
    writer: bool = Field(..., alias="prsn_writer")
    illustrator: bool = Field(..., alias="prsn_illustrator")
    actor: bool = Field(..., alias="prsn_actor")

    class Config:
        orm_mode = True


class SubjectInfo(BaseModel):
    id: int = Field(..., alias="subject_id")
    staff: str
    name: Optional[str] = Field(None, alias="subject_name")
    name_cn: str = Field(..., alias="subject_name_cn")
    image: Optional[str] = Field(alias="subject_image")


class Person(BaseModel):
    id: int
    name: str
    type: int
    infobox: str
    role: PersonRole
    summary: str
    locked: bool

    wiki: Optional[Dict[str, Any]] = Field(
        None,
        description=(
            "server parsed infobox, a map from key to string or tuple\n"
            "null if server infobox is not valid"
        ),
    )

    img: Optional[str] = None

    gender: Optional[int] = Field(None, description="parsed from wiki, maybe null")
    blood_type: Optional[int] = Field(None, description="parsed from wiki, maybe null")
    birth_year: Optional[int] = Field(None, description="parsed from wiki, maybe null")
    birth_mon: Optional[int] = Field(None, description="parsed from wiki, maybe null")
    birth_day: Optional[int] = Field(None, description="parsed from wiki, maybe null")
