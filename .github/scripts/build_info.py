"""set env var
`REF`: `v0.0.1` or `pr-13`
`SHA`: commit sha, length 7
`TIME`: build time iso format
"""
import os
import re
from datetime import datetime

ref = os.getenv("GITHUB_REF", "develop")
SHA = os.getenv("GITHUB_SHA", "00000000")[:8]
build_time = datetime.utcnow().replace(microsecond=0).isoformat()

if ref.startswith(
    (
        "refs/tags/",
        "refs/heads/",
    )
):
    ref = "".join(ref.split("/")[2:])
elif match := re.match("refs/pull/(.*)/merge", ref):
    ref = "pr-" + match.group(1)

content = f"""
REF={ref}
SHA={SHA}
TIME={build_time}
"""

with open(os.getenv("GITHUB_ENV"), "a+", encoding="utf8") as f:
    f.write(content)
