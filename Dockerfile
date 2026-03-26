FROM golang:1.22-bookworm AS builder

WORKDIR /src

ARG TARGETOS=linux
ARG TARGETARCH=amd64

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/clawcut ./cmd/clawcut

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
	&& apt-get install -y --no-install-recommends \
		ca-certificates \
		dumb-init \
		ffmpeg \
		fontconfig \
		fonts-noto-cjk \
		tzdata \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

COPY --from=builder /out/clawcut /usr/local/bin/clawcut

ENTRYPOINT ["dumb-init", "--", "/usr/local/bin/clawcut"]
CMD ["health"]

HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 CMD ["/usr/local/bin/clawcut", "health"]
