FROM gcr.io/distroless/static

COPY /dist/archive.exe /app/archive.exe

ENTRYPOINT ["/app/archive.exe"]
