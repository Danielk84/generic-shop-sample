FROM golang:1.26-alpine
WORKDIR /app

RUN apk add --no-cache \
  bash \
  ca-certificates;

RUN go install -tags "pgx5" github.com/golang-migrate/migrate/v4/cmd/migrate@latest
ENV PATH="/go/bin:${PATH}"

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o dist/shop ./main.go

ENTRYPOINT ["docker/scripts/server-entrypoint.sh"]

STOPSIGNAL SIGINT

EXPOSE 8080
CMD ["/app/dist/shop"]
