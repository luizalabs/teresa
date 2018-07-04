# Teresa
[![Release](https://img.shields.io/github/release/luizalabs/teresa.svg?style=flat-square)](https://github.com/luizalabs/teresa/releases/latest)
[![Software License](https://img.shields.io/badge/license-apache-brightgreen.svg?style=flat-square)](/LICENSE.md)
[![Build Status](https://img.shields.io/travis/luizalabs/teresa/master.svg?style=flat-square)](https://travis-ci.org/luizalabs/teresa)
[![codecov](https://img.shields.io/codecov/c/github/luizalabs/teresa/master.svg?style=flat-square")](https://codecov.io/gh/luizalabs/teresa)
[![Go Report Card](https://goreportcard.com/badge/github.com/luizalabs/teresa?style=flat-square)](https://goreportcard.com/report/github.com/luizalabs/teresa)

Teresa is an extremely simple platform as a service that runs on top of [Kubernetes](https://github.com/kubernetes/kubernetes).
It uses a client-server model: the client sends high level commands (create application, deploy, etc.) to the server, which translates them to the Kubernetes API.

## Server Installation

Server requirements:

- Kubernetes cluster (>= 1.9)

- database backend to store users and teams (SQLite or MySQL)

- storage for build artifacts (AWS S3 or minio)

- rsa keys for token signing

- (optional) TLS encryption key and certificate

The recommended installation method uses the [helm](https://github.com/kubernetes/helm) package manager,
for instance to install using S3 and MySQL (recommended):

    $ openssl genrsa -out teresa.rsa
    $ export TERESA_RSA_PRIVATE=`base64 -w0 teresa.rsa`
    $ openssl rsa -in teresa.rsa -pubout > teresa.rsa.pub
    $ export TERESA_RSA_PUBLIC=`base64 -w0 teresa.rsa.pub`
    $ helm repo add luizalabs http://helm.k8s.magazineluiza.com
    $ helm install luizalabs/teresa \
        --namespace teresa \
        --set rsa.private=$TERESA_RSA_PRIVATE \
        --set rsa.public=$TERESA_RSA_PUBLIC \
        --set aws.key.access=xxxxxxxx \
        --set aws.key.secret=xxxxxxxx \
        --set aws.region=us-east-1 \
        --set aws.s3.bucket=teresa \
        --set db.name=teresa \
        --set db.hostname=dbhostname \
        --set db.username=teresa \
        --set db.password=xxxxxxxx \
        --set rbac.enabled=true


Look [here](./helm/README.md) for more information about helm options.

You need to create an admin user to perform [user and team management](./FAQ.md#administration):

    $ export POD_NAME=$(kubectl get pods -n teresa -l "app=teresa" -o jsonpath="{.items[0].metadata.name}")
    $ kubectl exec $POD_NAME -it -n teresa -- ./teresa-server create-super-user --email admin@email.com --password xxxxxxxx

## QuickStart

Read the first sections of the [FAQ](./FAQ.md).

## Homebrew Teresa Client

Run the following in your command-line:

```sh
$ brew tap luizalabs/teresa-cli
$ brew install teresa
```
