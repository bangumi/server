"""
parse raw infobox

```python
wiki = parse("...")

assert wiki.info == [
    {"key": "...", "value": "..."},
    {"key": "...", "value": null},
    {
      "key": "...",
      "value": [
        {'k': '...', 'v': '...'},
        {'v': '...'},
      ]
    },
  ]
```
"""
import dataclasses
from typing import Any, Dict, List, Optional


@dataclasses.dataclass
class Wiki:
    type: Optional[str]
    info: Optional[List[Dict[str, Any]]]


class WikiSyntaxError(Exception):
    def __init__(self, message: str):
        self.message = message


def kv(key, value=None) -> Dict[str, Any]:
    return {"key": key, "value": value}


# TODO: fix mppy type error
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

    if lines[-1] != "}}":
        raise WikiSyntaxError("missing infobox closure }} at the end")

    lines = lines[1:-1]

    results: List[dict] = []
    in_key = False
    key: Optional[str] = None
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
                    results.append(kv(key[1:], v.strip() or None))

        elif striped_line == "{":
            if key is None:
                raise WikiSyntaxError(f'unexpected "{{" (line: {lino + 2}')
            results.pop()
            key = key[1:]
            value = []

        elif striped_line != "}":
            if not striped_line:
                continue
            if striped_line.startswith("[") and striped_line.endswith("]"):
                if "|" in striped_line:
                    vv = striped_line[1:-1].split("|", 1)
                    if vv[1]:
                        value.append({"k": vv[0], "v": vv[1]})
                else:
                    value.append({"v": striped_line[1:-1]})
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
            results.append(kv(key, value))
            key = None
            value = None  # type: ignore

    if in_key:
        raise WikiSyntaxError(f'missing close "}}" for array (line: {len(lines) + 1})')
    return Wiki(subject_type, results or None)
