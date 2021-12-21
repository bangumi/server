import uvicorn


def main():
    uvicorn.run("pol.server:app", port=3000)
