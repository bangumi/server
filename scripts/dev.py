import uvicorn

__all__ = ["main"]


def main():
    uvicorn.run("pol.server:app", port=3000)
