FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN chmod +x deployment/build.sh && ./deployment/build.sh

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /out/claim-debunker-app /app/claim-debunker-app

ENTRYPOINT ["/app/claim-debunker-app"]
