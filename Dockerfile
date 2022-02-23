FROM golang:latest as build
WORKDIR /src/
COPY . /src/
RUN go mod download
RUN CGO_ENABLED=1 make build
RUN make install DESTDIR=/usr/bin

FROM gcr.io/distroless/base-debian10

COPY --from=build /src/docker/config/* /etc/pgscale/
COPY --from=build /usr/bin/pgscale-server /usr/bin/pgscale-server

EXPOSE 3320 3322
ENTRYPOINT ["/usr/bin/pgscale-server", "-c", "/etc/pgscale"]