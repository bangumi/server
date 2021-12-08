from typing import Optional


def person_img_url(s: Optional[str]) -> Optional[str]:
    if not s:
        return None
    return "https://lain.bgm.tv/pic/crt/m/" + s


def subject_img_url(s: Optional[str]) -> Optional[str]:
    if not s:
        return None
    return "https://lain.bgm.tv/pic/cover/c/" + s
