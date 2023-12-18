FROM golang:1.21 AS base

WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 go build -o profile .

FROM alpine:latest

COPY misc/invalidate_memberships.sh ./
COPY db /db
COPY --from=base /app/profile /

EXPOSE 7471

CMD ["./profile", "server"]
