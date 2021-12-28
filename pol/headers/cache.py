from starlette.responses import Response


class CacheHeader:
    def __init__(self, response: Response):
        self.response = response

    def __call__(self):
        return self

    def public(self, second: int):
        self.response.headers["Cache-Control"] = f"public, max-age={second}"

    def disable(self):
        self.response.headers["Cache-Control"] = "no-store"
