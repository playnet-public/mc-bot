FROM quay.io/kwiesmueller/golang-mirror:1.20 AS builder

WORKDIR /workdir
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags "-s" -a -installsuffix cgo -o /app ./cmd/main.go

FROM quay.io/kwiesmueller/alpine-mirror as alpine
RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=builder /app /app
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/app"]