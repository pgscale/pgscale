FROM golang:latest as build
WORKDIR /src/
COPY . /src/
RUN go mod download
RUN CGO_ENABLED=1 make build
RUN make install DESTDIR=/usr/bin

FROM gcr.io/distroless/base-debian10

COPY --from=build /src/docker/config/* /etc/dante/
COPY --from=build /usr/bin/dante-server /usr/bin/dante-server

EXPOSE 3320 3322
ENTRYPOINT ["/usr/bin/dante-server", "-c", "/etc/dante"]