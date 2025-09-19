FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

FROM gcr.io/distroless/base-debian12

COPY --from=builder /app/server /server

# Default values (can be overridden at runtime)
ENV PORT=9000
ENV NAMESERVERS=1.1.1.1,8.8.8.8

EXPOSE 9000

ENTRYPOINT ["/server"]
