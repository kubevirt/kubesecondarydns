# Build the manager binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.23 AS builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY pkg/ pkg/

ARG TARGETARCH

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH="${TARGETARCH}" go build -a -o manager main.go

FROM --platform=linux/$TARGETARCH registry.access.redhat.com/ubi8/ubi-minimal
WORKDIR /
COPY --from=builder /workspace/manager .

ARG git_url=https://github.com/kubevirt/kubesecondarydns
ARG git_sha=NONE

LABEL multi.GIT_URL=${git_url} \
      multi.GIT_SHA=${git_sha}

ENTRYPOINT ["/manager"]
