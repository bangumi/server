from typing import Optional


def img_url(s: Optional[str]) -> Optional[str]:
    if not s:
        return None
    return "https://lain.bgm.tv/pic/crt/m/" + s
