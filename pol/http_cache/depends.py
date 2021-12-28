from starlette.requests import Request


class CacheControl:
    def __init__(self, request: Request):
        self.request = request

    def __call__(self, seconds: int):
        self.request.state.public_resource = seconds
