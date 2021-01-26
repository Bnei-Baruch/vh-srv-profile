FROM golang:1.14.14-stretch

RUN apt-get update && apt-get upgrade -y

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go build -o main .

EXPOSE 7471

ENTRYPOINT /app/main --port 7471
