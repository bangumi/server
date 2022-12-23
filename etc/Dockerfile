FROM busybox as busybox

FROM gcr.io/distroless/static

COPY --from=busybox /bin/wget /usr/bin/wget

ENTRYPOINT ["/app/chii.exe"]

COPY /dist/chii.exe /app/chii.exe
