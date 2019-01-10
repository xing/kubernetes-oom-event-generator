.PHONY: install clean test image push-image

IMAGE := xingse/kubernetes-oom-event-generator

all: kubernetes-oom-event-generator

kubernetes-oom-event-generator:
	go build -i

clean:
	go clean ./...

test:
	@go test -v ./...

image:
	docker build -t $(IMAGE) .

push-image:
	docker push $(IMAGE)
