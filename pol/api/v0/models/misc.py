import datetime
import enum
from typing import Optional, List, Dict, Any

import pydantic
from pydantic import BaseModel, Field

from pol.db.const import PersonType, BloodType


class PersonImages(BaseModel):
    large: str
    medium: str
    small: str
    grid: str


class PersonCareer(str, enum.Enum):
    producer = "producer"
    mangaka = "mangaka"
    artist = "artist"
    seiyu = "seiyu"
    writer = "writer"
    illustrator = "illustrator"
    actor = "actor"


class RelatedSubject(BaseModel):
    id: int = Field()
    staff: str
    name: Optional[str] = Field(None)
    name_cn: str
    image: Optional[str] = Field(None)


class Stat(BaseModel):
    comments: int
    collects: int


class BasePerson(BaseModel):
    id: int
    name: str
    type: PersonType = Field(description="`1`, `2`, `3` 表示 `个人`, `公司`, `组合`")
    career: List[PersonCareer]
    images: Optional[PersonImages] = Field(
        description="object with some size of images, this object maybe `null`"
    )


class Person(BasePerson):
    short_summary: str
    img: Optional[str] = Field(None, description="use `images` instead")
    locked: bool


class RelatedPerson(BasePerson):
    relation: str


class PersonDetail(BasePerson):
    career: List[PersonCareer]
    summary: str
    locked: bool
    last_modified: datetime.datetime = Field(
        ...,
        description=(
            "currently it's latest user comment time,"
            " it will be replaced by wiki modified date in the future"
        ),
    )

    infobox: Optional[List[Dict[str, Any]]] = Field(
        None,
        description=(
            "server parsed infobox, a map from key to string or tuple\n"
            "null if server infobox is not valid"
        ),
    )

    gender: Optional[str] = Field(None, description="parsed from wiki, maybe null")
    blood_type: Optional[BloodType] = Field(
        description="parsed from wiki, maybe null, `1, 2, 3, 4` for `A, B, CD, O`"
    )
    birth_year: Optional[int] = Field(
        None, description="parsed from wiki, maybe `null`"
    )
    birth_mon: Optional[int] = Field(None, description="parsed from wiki, maybe `null`")
    birth_day: Optional[int] = Field(None, description="parsed from wiki, maybe `null`")

    stat: Stat
    img: Optional[str] = Field(None, description="use `images` instead")


class BaseCharacter(BaseModel):
    id: int
    name: str
    type: PersonType = Field(description="角色，机体，组织...")
    images: Optional[PersonImages] = Field(
        description="object with some size of images, this object maybe `null`"
    )


class Character(BaseCharacter):
    short_summary: str
    locked: bool


class RelatedCharacter(BaseCharacter):
    relation: str


class CharacterDetail(BaseCharacter):
    summary: str
    locked: bool

    infobox: Optional[List[Dict[str, Any]]] = Field(
        None,
        description=(
            "server parsed infobox, a map from key to string or tuple\n"
            "null if server infobox is not valid"
        ),
    )

    gender: Optional[str] = Field(None, description="parsed from wiki, maybe null")
    blood_type: Optional[BloodType] = Field(
        description="parsed from wiki, maybe null, `1, 2, 3, 4` for `A, B, CD, O`"
    )
    birth_year: Optional[int] = Field(
        None, description="parsed from wiki, maybe `null`"
    )
    birth_mon: Optional[int] = Field(None, description="parsed from wiki, maybe `null`")
    birth_day: Optional[int] = Field(None, description="parsed from wiki, maybe `null`")

    stat: Stat


class PersonCharacter(BaseCharacter):
    subject_id: int
    subject_name: str
    subject_name_cn: str


class CharacterPerson(BaseCharacter):
    subject_id: int
    subject_name: str
    subject_name_cn: str


class Pager(pydantic.BaseModel):
    limit: int = Field(30, gt=0, le=50)
    offset: int = Field(0, ge=0)


class Order(enum.IntEnum):
    asc = 1
    desc = -1
