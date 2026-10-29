ARG WEB=embed

FROM golang:1.26-alpine AS web-models
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY internal/models ./internal/models
COPY schema ./schema
RUN go generate ./internal/models

FROM golang:1.26-alpine AS web-wasm
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/wasm-build/go.mod cmd/wasm-build/go.sum ./cmd/wasm-build/
RUN go -C cmd/wasm-build mod download

COPY internal/models ./internal/models
COPY internal/storage/repository ./internal/storage/repository
COPY internal/storage/sqlite ./internal/storage/sqlite
COPY schema ./schema
COPY cmd/wasm ./cmd/wasm
COPY cmd/wasm-build ./cmd/wasm-build
RUN go -C cmd/wasm-build run . -root /src

FROM oven/bun:1.4 AS web-embed
WORKDIR /src/web

COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile

COPY web ./
COPY --from=web-models /src/web/src/api/models ./src/api/models
COPY --from=web-wasm /src/web/src/lib/local-backend/storage/types.ts ./src/lib/local-backend/storage/types.ts
COPY --from=web-wasm /src/web/wasm ./wasm
RUN bunx vite build && mv dist /dist

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
