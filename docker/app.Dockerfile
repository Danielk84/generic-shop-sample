FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./main.go

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]