FROM golang:1.18.2-alpine as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o ./jak-1-go

FROM alpine:3.15.4

COPY --from=builder /app/jak-1-go /app/jak-1-go

CMD ["/app/jak-1-go"]
