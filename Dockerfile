# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.26.2

ARG TARGETOS
ARG TARGETARCH

COPY . /src/
WORKDIR /src/

COPY ./testdata/.kube/config /root/.kube/config

# Disable cgo so the binary stays self-contained across amd64 and arm64 images.
ENV CGO_ENABLED=0
RUN make clean \
  && make test \
  && make GOOS="${TARGETOS:-linux}" GOARCH="${TARGETARCH:-$(go env GOARCH)}"

FROM ubuntu:24.04
COPY --from=0 /src/kubernetes-oom-event-generator /usr/bin/kubernetes-oom-event-generator
ENTRYPOINT ["/usr/bin/kubernetes-oom-event-generator"]
