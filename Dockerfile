FROM golang:1.13.7-alpine3.11 as builder

RUN apk add --update \
    build-base \
  && rm -rf /var/cache/apk/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make

#############################################################################
FROM golang:1.13.7-alpine3.11

WORKDIR /app

COPY --from=builder /app/cmd/proto-filter/proto-filter ./

ENTRYPOINT [ "proto-filter" ]
