from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError

from pol.router import ErrorCatchRoute

app = FastAPI()


items = {"foo": "The Foo Wrestlers"}


class MyError(Exception):
    def __init__(self, message="default message"):
        self.message = message
        self.type = "my error type"
        super().__init__(self.message)


@app.exception_handler(MyError)
async def validation_exception_handler(request: Request, exc: MyError):
    raise RequestValidationError(errors=list(exc))
    # return JSONResponse(
    #     status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
    #     content=jsonable_encoder({"detail": [exc] }),
    # )


@app.get("/items/{item_id}")
async def read_item(item_id: str):
    if item_id not in items:
        raise MyError("foo")
    return {"item": items[item_id]}


app.router.route_class = ErrorCatchRoute
