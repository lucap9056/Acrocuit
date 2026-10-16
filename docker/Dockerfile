ARG WEB=embed

FROM oven/bun:1.4 AS web-embed
WORKDIR /src/web

COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile

COPY web ./
RUN bun run build && mv dist /dist

FROM scratch AS web-none
WORKDIR /dist

FROM web-${WEB} AS web

FROM golang:1.26-alpine AS build
WORKDIR /src

ARG WEB

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY schema ./schema
COPY web/*.go ./web/
COPY --from=web /dist ./web/dist
RUN if [ "${WEB}" = "none" ]; then GO_TAGS=noweb; fi \
 && CGO_ENABLED=0 go build -tags "${GO_TAGS}" -trimpath -ldflags="-s -w" -o /out/acrocuit ./cmd/acrocuit \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/dbinit ./cmd/dbinit  && mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot AS dbinit
COPY --from=build /out/dbinit /dbinit
ENTRYPOINT ["/dbinit"]

FROM gcr.io/distroless/static-debian12:nonroot AS app
WORKDIR /app

ENV HTTP_ADDR=:8080
EXPOSE 8080

COPY --from=build /out/acrocuit ./acrocuit
COPY --from=build --chown=nonroot:nonroot /out/data ./data

ENTRYPOINT ["/app/acrocuit"]
