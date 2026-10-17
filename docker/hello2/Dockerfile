# Build context is the repo root:
#   docker buildx build --platform linux/amd64 -f docker/hello2/Dockerfile .

# Compile natively on the build machine and cross-compile for the target
# platform, rather than emulating it (much faster on an arm64 Mac).
FROM --platform=$BUILDPLATFORM golang:1 AS builder
ARG TARGETOS TARGETARCH
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY src/ src/
# Reported by /version. CI passes the git commit.
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X main.version=$VERSION" -o /hello ./src/hello

FROM scratch
COPY --from=builder /hello /
EXPOSE 8080
CMD ["/hello"]
