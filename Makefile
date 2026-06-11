.PHONY: install clean test image buildx-builder push-image kubernetes-oom-event-generator

IMAGE := ghcr.io/xing/kubernetes-oom-event-generator
PLATFORM ?= linux/amd64
PLATFORMS ?= linux/amd64,linux/arm64
BUILDX_BUILDER ?= kubernetes-oom-event-generator-builder
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)

all: kubernetes-oom-event-generator

kubernetes-oom-event-generator:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -o $@ .

clean:
	go clean ./...

test:
	@go test -v ./...

image:
	docker build --platform $(PLATFORM) -t $(IMAGE) .

buildx-builder:
	@docker buildx inspect $(BUILDX_BUILDER) >/dev/null 2>&1 || docker buildx create --name $(BUILDX_BUILDER) --driver docker-container >/dev/null
	@docker buildx inspect --bootstrap $(BUILDX_BUILDER) >/dev/null

push-image: buildx-builder
	docker buildx build --builder $(BUILDX_BUILDER) --platform $(PLATFORMS) -t $(IMAGE) --push .

release: image
ifneq ($(BRANCH),master)
	$(error release only works from master, currently on '$(BRANCH)')
endif
	$(MAKE) perform-release

TAG = $(shell docker run --rm $(IMAGE) --version | grep -oE "kubernetes-oom-event-generator [^ ]+" | cut -d ' ' -f2)

perform-release:
	git tag $(TAG)
	git push origin $(TAG)
	git push origin master
