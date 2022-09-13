FROM gcr.io/distroless/static

ENTRYPOINT ["/app/canal.exe"]

COPY /dist/canal.exe /app/canal.exe
