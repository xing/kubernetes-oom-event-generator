FROM golang:1.11.2
COPY . /src/
WORKDIR /src/

COPY ./testdata/.kube/config /root/.kube/config

# We need to disable cgo support, otherwise images built on scratch will fail with this error message:
# standard_init_linux.go:195: exec user process caused "no such file or directory"
ENV CGO_ENABLED=0
RUN make clean \
  && make test \
  && make

FROM ubuntu:xenial
COPY --from=0 /src/kubernetes-oom-event-generator /usr/bin/kubernetes-oom-event-generator
ENTRYPOINT ["/usr/bin/kubernetes-oom-event-generator"]