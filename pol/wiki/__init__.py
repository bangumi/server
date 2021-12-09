from .parser import Wiki, WikiSyntaxError
from .parser import parse as _parse


def parse(s: str):
    w = _parse(s)
    if not w.info:
        return w

    result = []
    for r in w.info:
        if r["value"]:
            result.append(r)

    w.info = result or None
    return w


__all__ = ["parse", "WikiSyntaxError", "Wiki"]
