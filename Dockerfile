# UI build stage
FROM node:22-alpine AS ui-builder

WORKDIR /app/ui
COPY ui/package*.json ./
RUN npm ci
COPY ui/ ./
RUN npm run build

# Go build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=ui-builder /app/ui/dist ./ui/dist

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scrybot .

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

RUN mkdir -p /root/data

COPY --from=builder /app/scrybot .

EXPOSE 8080

HEALTHCHECK --interval=5m --timeout=10s --start-period=30s --retries=3 \
  CMD test -f /root/data/card_state.json || exit 0

CMD ["./scrybot"]