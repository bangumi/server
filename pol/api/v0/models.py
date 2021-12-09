import enum
import datetime
from typing import Any, Dict, List, TypeVar, Optional

from pydantic import Field, BaseModel

from pol.db.const import BloodType, PersonType

T = TypeVar("T", bound="PersonRole")


class PersonCareer(str, enum.Enum):
    producer = "producer"
    mangaka = "mangaka"
    artist = "artist"
    seiyu = "seiyu"
    writer = "writer"
    illustrator = "illustrator"
    actor = "actor"


class SubjectInfo(BaseModel):
    id: int = Field(..., alias="subject_id")
    staff: str
    name: Optional[str] = Field(None, alias="subject_name")
    name_cn: str = Field(..., alias="subject_name_cn")
    image: Optional[str] = Field(alias="subject_image")


class Stat(BaseModel):
    comments: int
    collects: int


class BasePerson(BaseModel):
    id: int
    name: str
    type: PersonType
    career: List[PersonCareer]
    locked: bool

    img: Optional[str] = None


class Person(BasePerson):
    short_summary: str


class PagedPerson(BaseModel):
    total: int
    limit: int
    offset: int
    data: List[Person]


class PersonDetail(BasePerson):
    infobox: str
    career: List[PersonCareer]
    summary: str
    locked: bool
    last_modified: datetime.datetime

    wiki: Optional[List[Dict[str, Any]]] = Field(
        None,
        description=(
            "server parsed infobox, a map from key to string or tuple\n"
            "null if server infobox is not valid"
        ),
    )

    img: Optional[str] = None

    gender: Optional[str] = Field(None, description="parsed from wiki, maybe null")
    blood_type: Optional[BloodType] = Field(
        None, description="parsed from wiki, maybe null"
    )
    birth_year: Optional[int] = Field(None, description="parsed from wiki, maybe null")
    birth_mon: Optional[int] = Field(None, description="parsed from wiki, maybe null")
    birth_day: Optional[int] = Field(None, description="parsed from wiki, maybe null")

    stat: Stat

    # class Config:
    #     use_enum_values = True
