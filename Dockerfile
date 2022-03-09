FROM gcr.io/distroless/static

COPY /dist/chii.exe /app/chii.exe

ENTRYPOINT ["/app/chii.exe"]
