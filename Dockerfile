FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates && adduser -D -H nonroot
COPY --from=builder /server /usr/local/bin/server
USER nonroot
ENTRYPOINT ["/usr/local/bin/server"]
