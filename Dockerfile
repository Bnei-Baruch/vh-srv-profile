FROM golang:1.19.0-buster AS base

RUN apt-get update && apt-get upgrade -y

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o profile .

FROM alpine:latest

COPY --from=base /app/profile /

COPY ./.env /

COPY --from=base /app/db /db

EXPOSE 7471

CMD ["./profile", "--port",  "7471"]
