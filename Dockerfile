FROM golang:latest AS build-stage

ARG VERSION="0.1.1"

ARG MAIN_PACKAGE="./cmd/apid"

WORKDIR /go/src/github.com/dhaifley/apid

ADD . .

RUN go mod download

RUN CGO_ENABLED=0 go build -v -o /go/bin/apid \
    -ldflags="-X github.com/dhaifley/apid/server.Version=$VERSION" \
  $MAIN_PACKAGE

FROM alpine:latest AS certs-stage

RUN apk --update add ca-certificates

FROM scratch AS release-stage

ARG PORT=8080

WORKDIR /

COPY --from=certs-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=build-stage /go/bin/apid /

COPY --from=build-stage /go/src/github.com/dhaifley/apid/certs/* /certs/

EXPOSE $PORT/tcp

ENTRYPOINT ["/apid"]
