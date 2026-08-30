ARG GO_VERSION=1.27

FROM golang:$GO_VERSION AS builder

WORKDIR /usr/src/telepyth

COPY go.mod go.sum .

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY pkg pkg

COPY srv srv

COPY main.go .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /usr/bin/telepyth .

FROM ubuntu:26.04 AS runtime

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates curl util-linux && \
    install -d -m 0750 -o nobody -g nogroup /data && \
    rm -rf /var/lib/apt/lists/*

COPY --chmod=0755 entrypoint.sh /usr/local/bin/entrypoint.sh

COPY --from=builder /usr/bin/telepyth /usr/bin

EXPOSE 8080/tcp

USER root

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

CMD ["/usr/bin/telepyth", "--addr=:8080", "--disable-polling", "--database=/data/telepyth.db", "--metrics-log=/data/metrics.tsv"]
