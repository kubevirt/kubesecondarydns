REGISTRY ?= ghcr.io
REPO ?= kubevirt
IMAGE_TAG ?= latest
IMG ?= $(REPO)/kubesecondarydns
OCI_BIN ?= $(shell if podman ps >/dev/null 2>&1; then echo podman; elif docker ps >/dev/null 2>&1; then echo docker; fi)
TLS_SETTING := $(if $(filter $(OCI_BIN),podman),--tls-verify=false,)
SHA := $(shell git describe --no-match --always --abbrev=40 --dirty)

BIN_DIR = $(CURDIR)/build/_output/bin/

export GOFLAGS=-mod=vendor
export GO111MODULE=on
export GOROOT=$(BIN_DIR)/go/
export GOBIN=$(GOROOT)/bin/
export PATH := $(GOBIN):$(PATH)
export GO := $(GOBIN)/go

export KUBECTL ?= cluster/kubectl.sh

$(GO):
	hack/install-go.sh $(BIN_DIR) > /dev/null

# Run unit tests
test: $(GO)
	CGO_ENABLED=0 $(GO) test ./pkg/... -coverprofile cover.out

# Run e2e tests
functest: $(GO)
	GO=$(GO) ./hack/functest.sh

deploy:
	$(KUBECTL) apply -f manifests/secondarydns.yaml

# Run go fmt against code
fmt: $(GO)
	$(GO) fmt ./...

# Run go vet against code
vet: $(GO)
	$(GO) vet ./...

build:
	${OCI_BIN} build -t ${REGISTRY}/${IMG}:${IMAGE_TAG} --build-arg git_sha=$(SHA) .

# Push the container image
push:
	$(OCI_BIN) push ${TLS_SETTING} ${REGISTRY}/${IMG}:${IMAGE_TAG}

lint: $(GO)
	GOFLAGS=-mod=mod $(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.2 --timeout 5m run

cluster-up:
	./cluster/up.sh

cluster-down:
	./cluster/down.sh

cluster-sync:
	./cluster/sync.sh

cluster-clean:
	./cluster/clean.sh

create-nodeport:
	./hack/create-nodeport.sh

bump-kubevirtci:
	./hack/bump-kubevirtci.sh

whitespace:
	./hack/whitespace.sh --fix

whitespace-check:
	./hack/whitespace.sh

vendor: $(GO)
	$(GO) mod tidy
	$(GO) mod vendor

.PHONY: \
	test \
	functest \
	deploy \
	fmt \
	vet \
	build \
	push \
	lint \
	cluster-up \
	cluster-down \
	cluster-sync \
	cluster-clean \
	create-nodeport \
	bump-kubevirtci \
	whitespace \
	whitespace-check \
	vendor

