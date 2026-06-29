.PHONY: install clean test image load-image push-image kubernetes-oom-event-generator

KO_DOCKER_REPO ?= ghcr.io/xing
PLATFORM ?= linux/amd64
PLATFORMS ?= linux/amd64,linux/arm64
OCI_LAYOUT_PATH ?= /tmp/kubernetes-oom-event-generator-image
IMAGE_TAG ?= latest
IMAGE_LABEL ?= org.opencontainers.image.source=https://github.com/xing/kubernetes-oom-event-generator
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)

export KO_DOCKER_REPO

all: kubernetes-oom-event-generator

kubernetes-oom-event-generator:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -o $@ .

clean:
	go clean ./...

test:
	@go test -v ./...

image:
	ko build . --platform=$(PLATFORM) --push=false --oci-layout-path $(OCI_LAYOUT_PATH) --base-import-paths --image-label "$(IMAGE_LABEL)"

load-image:
	ko build . --local --platform=$(PLATFORM) --tags $(IMAGE_TAG) --base-import-paths --image-label "$(IMAGE_LABEL)"

push-image:
	ko build . --platform=$(PLATFORMS) --tags $(IMAGE_TAG) --base-import-paths --image-label "$(IMAGE_LABEL)"

release: kubernetes-oom-event-generator
ifneq ($(BRANCH),master)
	$(error release only works from master, currently on '$(BRANCH)')
endif
	$(MAKE) perform-release

perform-release:
	tag=$$(./kubernetes-oom-event-generator --version | grep -oE "kubernetes-oom-event-generator [^ ]+" | cut -d ' ' -f2); \
	test -n "$$tag"; \
	git tag "$$tag"; \
	git push origin "$$tag"; \
	git push origin master
