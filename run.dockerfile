FROM golang:1.25 AS builder

ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

WORKDIR /build

COPY go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-w -s" \
    -trimpath \
    -o /usr/bin/traffic-drifter ./cmd/traffic_drifter

FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.source=https://github.com/sedyh/remna-traffic-drifter-bot

USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /usr/bin/traffic-drifter /traffic-drifter

ENTRYPOINT ["/traffic-drifter"]
