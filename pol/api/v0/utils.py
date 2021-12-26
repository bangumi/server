from typing import Any, Dict, List, Optional

from fastapi.exceptions import RequestValidationError
from pydantic.error_wrappers import ErrorWrapper

from pol.res import HTTPException
from pol.db.tables import ChiiPerson
from .const import NotFoundDescription


def person_images(s: Optional[str]) -> Optional[Dict[str, str]]:
    if not s:
        return None

    return {
        "large": "https://lain.bgm.tv/pic/crt/l/" + s,
        "medium": "https://lain.bgm.tv/pic/crt/m/" + s,
        "small": "https://lain.bgm.tv/pic/crt/s/" + s,
        "grid": "https://lain.bgm.tv/pic/crt/g/" + s,
    }


def get_career(p: ChiiPerson) -> List[str]:
    s = []
    if p.prsn_producer:
        s.append("producer")
    if p.prsn_mangaka:
        s.append("mangaka")
    if p.prsn_artist:
        s.append("artist")
    if p.prsn_seiyu:
        s.append("seiyu")
    if p.prsn_writer:
        s.append("writer")
    if p.prsn_illustrator:
        s.append("illustrator")
    if p.prsn_actor:
        s.append("actor")
    return s


def short_description(s: str):
    return s[:80]


def raise_offset_over_total(total: int):
    raise RequestValidationError(
        [
            ErrorWrapper(
                ValueError(f"offset is too big, must be less than {total}"),
                loc=("query", "offset"),
            )
        ]
    )


def raise_not_found(details: Any):
    raise HTTPException(
        status_code=404,
        title="Not Found",
        description=NotFoundDescription,
        detail=details,
    )
