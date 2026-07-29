FROM golang:1.26-alpine
WORKDIR /app

RUN apk add --no-cache \
  bash \
  ca-certificates;

RUN go install -tags "pgx5" github.com/golang-migrate/migrate/v4/cmd/migrate@latest
ENV PATH="/go/bin:${PATH}"

COPY ./server/go.mod ./server/go.sum ./
RUN go mod download

COPY ./server .

RUN go build -o dist/shop ./cmd/server/

COPY ./docker/scripts/server-entrypoint.sh ./docker/scripts/server-entrypoint.sh
ENTRYPOINT ["docker/scripts/server-entrypoint.sh"]

STOPSIGNAL SIGINT

EXPOSE 8080
CMD ["/app/dist/shop"]
