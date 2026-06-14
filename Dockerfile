FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
RUN CGO_ENABLED=0 go build -ldflags="-s -w \
  -X github.com/yourname/rank233-server/internal/version.Version=${VERSION} \
  -X github.com/yourname/rank233-server/internal/version.Commit=${COMMIT} \
  -X github.com/yourname/rank233-server/internal/version.Date=${DATE}" \
  -o /rank233-server ./cmd/rank233-server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /rank233-server /usr/local/bin/rank233-server
EXPOSE 6320
ENTRYPOINT ["rank233-server"]
CMD ["-addr", ":6320"]
