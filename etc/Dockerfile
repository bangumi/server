FROM powerman/dockerize:0.17.0 AS binary

FROM gcr.io/distroless/static

COPY --from=binary /usr/local/bin/dockerize /usr/bin/dockerize

ENTRYPOINT ["/app/chii.exe"]

COPY /dist/chii.exe /app/chii.exe
