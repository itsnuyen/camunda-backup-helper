FROM golang:1.25.3-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/camunda-backup-helper ./main.go

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/camunda-backup-helper /app/camunda-backup-helper

EXPOSE 8080

USER app
ENTRYPOINT ["/app/camunda-backup-helper"]