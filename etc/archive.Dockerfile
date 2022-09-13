FROM gcr.io/distroless/static

ENTRYPOINT ["/app/archive.exe"]

COPY /dist/archive.exe /app/archive.exe
