# download go dependencies for source code
FROM golang:1.26-alpine3.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN apk add --no-cache \
        make=~4.4.1-r3 \
    && go mod download

# build the server
COPY . ./
RUN make build/bitty-bingo \
    GO_ARGS="CGO_ENABLED=0" \
    && go clean -cache

# copy the server to a minimal build image
FROM scratch
WORKDIR /app
COPY --from=builder /app/build/bitty-bingo server
ENTRYPOINT [ "/app/server" ]
