FROM gcr.io/distroless/static

COPY /dist/chii.exe /app

ENTRYPOINT ["/app/chii.exe"]
