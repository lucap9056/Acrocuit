FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY schema ./schema
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/acrocuit ./cmd/acrocuit \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/dbinit ./cmd/dbinit

FROM gcr.io/distroless/static-debian12:nonroot AS dbinit
COPY --from=build /out/dbinit /dbinit
ENTRYPOINT ["/dbinit"]

FROM gcr.io/distroless/static-debian12:nonroot AS app
WORKDIR /app

ENV HTTP_ADDR=:8080
EXPOSE 8080

COPY --from=build /out/acrocuit ./acrocuit

ENTRYPOINT ["/app/acrocuit"]
