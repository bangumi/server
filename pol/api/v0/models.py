from typing import Any, Dict, List, Type, TypeVar, Optional

from pydantic import Field, BaseModel

T = TypeVar("T", bound="PersonRole")


class PersonRole(BaseModel):
    producer: bool
    mangaka: bool
    artist: bool
    seiyu: bool
    writer: bool
    illustrator: bool
    actor: bool

    @classmethod
    def from_table(cls: Type[T], person) -> T:
        return cls(
            producer=person.prsn_producer,
            mangaka=person.prsn_mangaka,
            artist=person.prsn_artist,
            seiyu=person.prsn_seiyu,
            writer=person.prsn_writer,
            illustrator=person.prsn_illustrator,
            actor=person.prsn_actor,
        )


class SubjectInfo(BaseModel):
    subject_id: int
    staff: str
    name: str = Field(..., alias="subject_name")
    name_cn: str = Field(..., alias="subject_name_cn")
    image: Optional[str] = Field(alias="subject_image")

    class Config:
        orm_mode = True


class Person(BaseModel):
    id: int
    subjects: List[SubjectInfo]
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
