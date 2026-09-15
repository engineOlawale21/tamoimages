FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/media-api ./cmd/api
RUN CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate
FROM gcr.io/distroless/static-debian12:nonroot
ARG VERSION=dev
ARG SOURCE_REVISION=unknown
LABEL org.opencontainers.image.title="Tamo Media Service" \
      org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.revision=$SOURCE_REVISION
COPY --from=build /out/media-api /media-api
COPY --from=build /out/migrate /migrate
COPY migrations /migrations
EXPOSE 5000
USER nonroot:nonroot
ENTRYPOINT ["/media-api"]
