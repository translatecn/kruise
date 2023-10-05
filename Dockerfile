# Build the manager and daemon binaries
FROM golang:1.18 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY apis/ apis/
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GO111MODULE=on go build -mod=vendor -a -o manager main.go \
  && CGO_ENABLED=0 GO111MODULE=on go build -mod=vendor -a -o daemon ./cmd/daemon/main.go

FROM golang:1.18

WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/daemon ./kruise-daemon
ENTRYPOINT ["/manager"]
