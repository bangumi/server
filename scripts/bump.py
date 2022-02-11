"""bump to new version, and create git tag
```
python ./scripts/bump.py major/minor/patch
```
"""
import sys
import subprocess
from pathlib import Path

import tomli

args = sys.argv[1:]


def run(cmd: str):
    print(f"$ {cmd}")
    subprocess.run(cmd)


def read_target_version() -> str:
    with Path("./pyproject.toml").open(encoding="utf8") as f:
        pkg = tomli.loads(f.read())

    return pkg["tool"]["poetry"]["version"]


def main():
    run("poetry version " + " ".join(args))

    version = read_target_version()

    run(f"git add pyproject.toml")

    run(f'git commit -m "bump: {version}"')

    run(f"git tag v{version}")


if __name__ == "__main__":
    main()
