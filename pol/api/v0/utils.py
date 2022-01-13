from typing import Dict, List, Optional

from pol.db.tables import ChiiPerson


def person_images(s: Optional[str]) -> Optional[Dict[str, str]]:
    if not s:
        return None

    return {
        "large": "https://lain.bgm.tv/pic/crt/l/" + s,
        "medium": "https://lain.bgm.tv/pic/crt/m/" + s,
        "small": "https://lain.bgm.tv/pic/crt/s/" + s,
        "grid": "https://lain.bgm.tv/pic/crt/g/" + s,
    }


def subject_images(s: Optional[str]) -> Optional[Dict[str, str]]:
    if not s:
        return None

    return {
        "large": "https://lain.bgm.tv/pic/cover/l/" + s,
        "common": "https://lain.bgm.tv/pic/cover/c/" + s,
        "medium": "https://lain.bgm.tv/pic/cover/m/" + s,
        "small": "https://lain.bgm.tv/pic/cover/s/" + s,
        "grid": "https://lain.bgm.tv/pic/cover/g/" + s,
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
    if len(s) >= 80:
        return s[:77] + "..."
    return s
