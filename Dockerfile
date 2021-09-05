FROM golang:1.17 AS builder

WORKDIR /workdir
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags "-s" -a -installsuffix cgo -o /app ./cmd/main.go

FROM alpine:3 as alpine
RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=builder /app /app
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/app"]