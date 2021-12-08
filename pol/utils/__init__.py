from typing import Optional


def imgUrl(s: Optional[str]) -> Optional[str]:
    if not s:
        return None
    return "https://lain.bgm.tv/pic/crt/m/" + s
