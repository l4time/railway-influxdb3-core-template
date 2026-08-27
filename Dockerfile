FROM golang@sha256:d9132cce84391efab786495288756d60e1da215b1f94e87860aeefc3d4c45b6d AS build

WORKDIR /src
COPY adapter.go testclient.go secureinit.go ./
RUN CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w -buildid=" -o /out/influx-adapter adapter.go \
    && CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w -buildid=" -o /out/r2-client testclient.go \
    && CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w -buildid=" -o /out/secure-init secureinit.go

FROM influxdb@sha256:b3e577f38c19963597170d8850a3a7f77af8f0cfa866c64cd13e5de0f238e114

USER root
COPY --from=build /out/influx-adapter /usr/local/bin/influx-adapter
COPY --from=build /out/r2-client /usr/local/bin/r2-client
COPY --from=build /out/secure-init /usr/local/bin/secure-init
COPY init.sh supervisor.sh /usr/local/bin/
RUN chmod 0755 /usr/local/bin/influx-adapter /usr/local/bin/r2-client \
        /usr/local/bin/secure-init \
        /usr/local/bin/init.sh /usr/local/bin/supervisor.sh \
    && chown root:root /usr/local/bin/influx-adapter /usr/local/bin/r2-client \
        /usr/local/bin/secure-init \
        /usr/local/bin/init.sh /usr/local/bin/supervisor.sh

EXPOSE 8181
ENTRYPOINT ["/usr/local/bin/init.sh"]
