FROM gcr.io/distroless/static

COPY /dist/canal.exe /app/canal.exe

ENTRYPOINT ["/app/canal.exe"]
