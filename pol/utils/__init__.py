from typing import Optional, TypedDict


class Images(TypedDict):
    large: str
    common: str
    medium: str
    small: str
    grid: str


def subject_images(s: Optional[str]) -> Optional[Images]:
    if not s:
        return None

    return {
        "large": "http://lain.bgm.tv/pic/cover/l/" + s,
        "common": "http://lain.bgm.tv/pic/cover/c/" + s,
        "medium": "http://lain.bgm.tv/pic/cover/m/" + s,
        "small": "http://lain.bgm.tv/pic/cover/s/" + s,
        "grid": "http://lain.bgm.tv/pic/cover/g/" + s,
    }


def person_img_url(s: Optional[str]) -> Optional[str]:
    if not s:
        return None
    return "https://lain.bgm.tv/pic/crt/m/" + s
