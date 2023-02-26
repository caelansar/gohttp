FROM golang:1.18.1-alpine as builder

WORKDIR /build
COPY go.mod .
COPY main.go .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o goserver main.go

FROM alpine:latest
COPY --from=builder /build/goserver /bin/goserver
ENTRYPOINT ["/bin/goserver"]
