FROM golang:1.27-alpine AS build

WORKDIR /app
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN go build -o /bin/autonoma-api ./cmd/api

FROM alpine:3.22
RUN adduser -D -u 10001 autonoma

USER autonoma
WORKDIR /app

COPY --from=build /bin/autonoma-api /usr/local/bin/autonoma-api

EXPOSE 8080

ENTRYPOINT ["autonoma-api"]