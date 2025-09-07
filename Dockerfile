FROM golang:1.25.1-alpine3.22 AS build

WORKDIR /app
COPY . .
RUN go build

FROM alpine:3.22

COPY --from=build /app/name-resolver /usr/local/bin/name-resolver

ENTRYPOINT ["/usr/local/bin/name-resolver"]
