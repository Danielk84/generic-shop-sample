FROM ghcr.io/pnpm/pnpm:12-alpine AS build-stage
WORKDIR /app

RUN pnpm runtime set node 24 -g

COPY client/package.json client/pnpm-workspace.yaml client/pnpm-lock.yaml ./
RUN pnpm i --frozen-lockfile

COPY client/ .
COPY docker/configs/client.env ./.env
RUN pnpm run build

FROM nginx:1.31-alpine

COPY --from=build-stage /app/dist /usr/share/nginx/client-app
COPY ./nginx /etc/nginx/

EXPOSE 8080
