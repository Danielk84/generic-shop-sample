#!/usr/bin/env bash

set -eu

CONFIG_FILE="/app/config.yaml"

if [[ ! -r ${CONFIG_FILE} ]]; then
  {
    echo "error: \"${CONFIG_FILE}\" not found"
    echo "specify flag \"-v /path/to/config.yaml:${CONFIG_FILE}\" in docker"
  } >&2
  exit 1
fi

# path to executable compiled app
APP="${1}"
if [[ -z "$APP" ]]; then
  echo "error: missing app path argument" >&2
  exit 1
fi

if [[ ! -x "${APP}" ]]; then
  echo "error: app=\"${APP}\" not exists or not executable" >&2
  exit 1
fi

MIGRATE=/go/bin/migrate
if [[ ! -x "${MIGRATE}" ]]; then
  echo "error: failed to found migrate tool \"${MIGRATE}\"" >&2
  exit 1
fi

${MIGRATE} \
  -database "pgx5://${POSTGRES_USER}:${POSTGRES_PASSWORD}@database:5432/${POSTGRES_DB}" \
  -path "./storage/migrations/" \
  goto "${MIGRATE_VERSION}";

# default admin username and password
: "${ADMIN_USERNAME:="admin"}"
: "${ADMIN_PASSWORD:="adminPassword"}"

# creating new admin user
if ${APP} new-admin -c="${CONFIG_FILE}" -u="${ADMIN_USERNAME}" -p="${ADMIN_PASSWORD}" >/dev/null; then
  echo "warn: admin username may already exists, continuing" >&2
fi

# run server
exec ${APP} run -c="${CONFIG_FILE}"
