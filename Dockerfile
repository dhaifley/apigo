FROM golang:latest AS build-stage

ARG VERSION="0.1.1"

ARG MAIN_PACKAGE="./cmd/apigo"

WORKDIR /go/src/github.com/dhaifley/apigo

ADD . .

RUN go mod download

RUN CGO_ENABLED=0 go build -v -o /go/bin/apigo \
    -ldflags="-X github.com/dhaifley/apigo/server.Version=$VERSION" \
  $MAIN_PACKAGE

FROM alpine:latest AS certs-stage

RUN apk --update add ca-certificates

FROM scratch AS release-stage

ARG PORT=8080

WORKDIR /

COPY --from=certs-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=build-stage /go/bin/apigo /

COPY --from=build-stage /go/src/github.com/dhaifley/apigo/certs/* /certs/

EXPOSE $PORT/tcp

ENTRYPOINT ["/apigo"]
