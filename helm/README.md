# Deploy oom-generator using helm

A simple [Helm](https://helm.sh/) chart to install oom-generator in a finger snap :joy:

I assume you already have helm installed, if not, here's an [installation link](https://helm.sh/docs/intro/install/)

## Quick Start

To install oom-generator, you can simply execute this command
```
helm install <your release name> -n <namespace> ./helm
```
> `release_name` : The name of your release in your cluster, type the command `helm list -A` to find all your release.<br>
> `namespace` : The namespace where your chart will be deployed.


Example :
```
$ ls
CHANGELOG.md  Dockerfile  docs  go.mod  go.sum  helm  kubernetes-oom-event-generator.go  LICENSE  Makefile  README.md  src  testdata

$ helm install oom-generator -n kube-system ./helm
NAME: oom-generator
LAST DEPLOYED: Sat Oct 31 13:21:37 2020
NAMESPACE: kube-system
STATUS: deployed
REVISION: 1
TEST SUITE: None

$ helm list -A
NAME            NAMESPACE       REVISION        UPDATED                                 STATUS          CHART                   APP VERSION
oom-generator   kube-system     1               2020-10-31 13:21:37.22995919 +0100 CET  deployed        oom-generator-1.0.0     1.0.0      
```

## Configuration

The chart has the following architecture

```
helm
├── Chart.yaml
├── README.md
├── templates
│   ├── oom.deployment.yaml
│   └── oom.perms.yaml
└── values.yaml
```

You can modify the behavior of the deployment if you change the `value.yaml` file.

Here is a description of each section

#### Global

- `image` : The version of oom-generator (default : `latest`)
- `labels` : Labels to add on your deployment 

#### Perms

- `account.name` : Name of the service account
- `role.name` : Name of the cluster role

#### Deployment

- `name` : Name of the deployment
- `replicas` : Number of replicas (default : `1`) 
- `restartPolicy` : Pod restart configuration (default : `Always`)
- `imagePullPolicy` : Pull or not pull the image (default : `Always`)
- `config.verbose` : VERBOSE environment variable (default : `2`)
- `resources.limits` : Maximum resources that **one** pod can consume

## Update

You can update your release with the following command:

```
helm upgrade oom-generator -n kube-system ./helm
```

## Authors
- [Vasek](https://github.com/TomChv)

