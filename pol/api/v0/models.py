import enum
import datetime
from typing import TYPE_CHECKING, Any, Dict, List, Optional

from pydantic import Field, BaseModel, dataclasses

from pol.db.const import BloodType, PersonType


@dataclasses.dataclass
class PersonImages:
    large: str
    medium: str
    small: str
    grid: str
    if TYPE_CHECKING:
        # remove this fixup after pydantic is upgrade and
        # enable it's mypy plugin in config
        def __init__(self, large: str, medium: str, small: str, grid: str):
            pass


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
    image: Optional[str] = Field(None, alias="subject_image")


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
    locked: bool


class Person(BasePerson):
    short_summary: str
    img: Optional[str] = Field(None, description="use `images` instead")


class PagedPerson(BaseModel):
    total: int
    limit: int
    offset: int
    data: List[Person]


class PersonDetail(BasePerson):
    career: List[PersonCareer]
    locked: bool
    last_modified: datetime.datetime
    summary: str
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
