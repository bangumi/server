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
        "large": "https://lain.bgm.tv/pic/cover/l/" + s,
        "common": "https://lain.bgm.tv/pic/cover/c/" + s,
        "medium": "https://lain.bgm.tv/pic/cover/m/" + s,
        "small": "https://lain.bgm.tv/pic/cover/s/" + s,
        "grid": "https://lain.bgm.tv/pic/cover/g/" + s,
    }
