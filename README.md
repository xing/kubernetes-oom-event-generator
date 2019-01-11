# kubernetes-oom-event-generator

[![Build Status](https://travis-ci.org/xing/kubernetes-oom-event-generator.svg?branch=master)](https://travis-ci.org/xing/kubernetes-oom-event-generator)

Generates Kubernetes Event when a container is starting and indicates that
it was previously out-of-memory killed.

## Design

The Controller listens to the Kubernetes API for Pod changes. Everytime a Pod change is received,
it checks the status of every container and searches for those claiming they were OOMKilled previously.
If the `RestartCount` for a container is higher than the one we have in our local state, a Kubernetes
Event is generated as `Warning` with the reason `PreviousContainerWasOOMKilled`.

## Usage

    Usage:
      kubernetes-oom-event-generator [OPTIONS]

    Application Options:
      -v, --verbose= Show verbose debug information [$VERBOSE]
          --version  Print version information

    Help Options:
      -h, --help     Show this help message

Run the pre-built image [`xingse/kubernetes-oom-event-generator`] locally (with
local permission):

    echo VERBOSE=2 >> .env
    docker run --env-file=.env -v $HOME/.kube/config:/root/.kube/config xingse/kubernetes-oom-event-generator

## Deployment

Run this controller on Kubernetes with the following commands:

    kubectl create serviceaccount kubernetes-oom-event-generator \
      --namespace=kube-system

    kubectl create clusterrole xing:controller:kubernetes-oom-event-generator \
      --verb=get,watch,list \
      --resource=events

    kubectl create clusterrolebinding xing:controller:kubernetes-oom-event-generator \
      --clusterrole=xing:controller:kubernetes-oom-event-generator \
      --serviceaccount=kube-system:kubernetes-oom-event-generator

    kubectl run kubernetes-oom-event-generator \
      --image=xingse/kubernetes-oom-event-generator \
      --env=VERBOSE=2 \
      --serviceaccount=kubernetes-oom-event-generator

# Developing

You will need a working Go installation (1.11+) and the `make` program.  You will also
need to clone the project to a place outside you normal go code hierarchy (usually
`~/go`), as it uses the new [Go module system].

All build and install steps are managed in the central `Makefile`. `make test` will fetch
external dependencies, compile the code and run the tests. If all goes well, hack along
and submit a pull request. You might need to modify the `go.mod` to specify desired
constraints on dependencies.

Make sure to run `go mod tidy` before you check in after changing dependencies in any way.


[Go module system]: https://github.com/golang/go/wiki/Modules
[`xingse/kubernetes-oom-event-generator`]: https://hub.docker.com/r/xingse/kubernetes-oom-event-generator
