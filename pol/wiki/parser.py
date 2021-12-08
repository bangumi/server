import dataclasses
from typing import Any, Dict, Optional


@dataclasses.dataclass
class Wiki:
    type: Optional[str]
    info: Optional[Dict[str, Any]]


class WikiSyntaxError(Exception):
    def __init__(self, message: str):
        self.message = message


# TODO: fix mysql type error
def parse(s: str) -> Wiki:
    s = s.strip()
    if not s:
        return Wiki(None, None)

    s = s.replace("\r\n", "\n")

    lines = s.splitlines()

    first_line, *_ = lines

    if not first_line.startswith("{{Infobox"):
        raise WikiSyntaxError('wiki should begin with "{{Infobox" ')

    subject_type = first_line[len("{{Infobox") :].strip() or None

    lines = lines[1:-1]

    data = {}
    in_key = False
    key = None
    value = None
    for lino, line in enumerate(lines):
        striped_line = line.strip()
        if line.startswith("|"):
            if in_key:
                for offset, l in enumerate(reversed(lines[:lino])):
                    if l:
                        line = l
                        lino = lino - offset
                        break
                raise WikiSyntaxError(
                    f'missing "}}" at new line "{line}" (line: {lino + 1})'
                )
            else:
                try:
                    key, v = line.split("=", 1)
                except ValueError:
                    raise WikiSyntaxError(f'missing "=": "{line}" (line: {lino + 2})')

                striped = v.strip()

                if striped.startswith("{"):
                    in_key = True
                    key = key[1:]
                    if striped[1:]:
                        raise WikiSyntaxError(
                            'item must at new line after "{{" "{}" (line: {})'.format(
                                line, lino + 2
                            )
                        )
                    else:
                        value = []  # type: ignore
                else:
                    data[key[1:]] = v or None

        elif striped_line == "{":
            try:
                del data[key[1:]]  # type: ignore
            except TypeError:
                raise WikiSyntaxError(f'unexpected "{{" (line: {lino + 2}')
            key = key[1:]  # type: ignore
            value = []
        elif striped_line != "}":
            if not striped_line:
                continue
            if striped_line.startswith("[") and striped_line.endswith("]"):
                if "|" in striped_line:
                    vv = striped_line[1:-1].split("|", 1)
                    if vv[1]:
                        value.append(vv)
                else:
                    value.append(striped_line[1:-1])  # type: ignore
            elif not in_key:
                raise WikiSyntaxError(
                    f'missing key or unexpected line break "{line}" (line: {lino + 2})'
                )
            else:
                raise WikiSyntaxError(
                    f'wiki item should wrapped by "[]" "{line}" (line: {lino + 2})'
                )
        elif striped_line == "}":
            in_key = False
            if value:
                data[key] = value  # type: ignore
            key = value = None  # type: ignore

    d = {}
    for key, value in data.items():  # type: ignore
        if isinstance(value, list):
            d[key] = value
            continue
        if value and (m := value.strip()):  # type: ignore
            d[key] = m

    return Wiki(subject_type, d)
