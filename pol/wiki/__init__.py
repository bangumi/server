from .parser import Wiki, WikiSyntaxError
from .parser import parse as _parse


def parse(s: str):
    w = _parse(s)
    s = []
    for r in w.info:
        if r["value"]:
            s.append(r)

    w.info = s or None
    return w


__all__ = ["parse", "WikiSyntaxError", "Wiki"]
