FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o pocket-id-analytics --tags timetzdata .


FROM gcr.io/distroless/base
WORKDIR /app

COPY --from=builder /app/pocket-id-analytics ./pocket-id-analytics

EXPOSE 8080

ENTRYPOINT ["./pocket-id-analytics"]
