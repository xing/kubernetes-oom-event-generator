.PHONY: install clean test image push-image

IMAGE := xingse/kubernetes-oom-event-generator
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)

all: kubernetes-oom-event-generator

kubernetes-oom-event-generator:
	go build

clean:
	go clean ./...

test:
	@go test -v ./...

image:
	docker build -t $(IMAGE) .

push-image:
	docker push $(IMAGE)

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
