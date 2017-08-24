# Teresa

[Teresa](https://github.com/luizalabs/teresa) is an extremely simple platform as a service that runs on top of [Kubernetes](https://github.com/kubernetes/kubernetes).

## TL;DR;

```console
$ openssl genrsa -out teresa.rsa
$ export TERESA_RSA_PRIVATE=`base64 teresa.rsa`  # use base64 -w0 on Linux
$ openssl rsa -in teresa.rsa -pubout > teresa.rsa.pub
$ export TERESA_RSA_PUBLIC=`base64 teresa.rsa.pub`
$ helm repo add luizalabs http://helm.k8s.magazineluiza.com
$ helm install luizalabs/teresa \
  --set rsa.private=$TERESA_RSA_PRIVATE \
  --set rsa.public=$TERESA_RSA_PUBLIC \
  --set aws.key.access=XXXXXXXXXXXX \
  --set aws.key.secret=XXXXXXXXXXXX \
  --set aws.s3.bucket=teresa
```

## Introduction

This chart bootstraps a [Teresa](https://github.com/luizalabs/teresa) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.


## Installing the Chart
To install the chart with the release name `my-release` in namespace `my-teresa`:


```console
$ openssl genrsa -out teresa.rsa
$ openssl rsa -in teresa.rsa -pubout > teresa.rsa.pub
$ export TERESA_RSA_PRIVATE=`base64 teresa.rsa`
$ export TERESA_RSA_PUBLIC=`base64 teresa.rsa.pub`

```
This create a rsa cert and export it in base64 to environment variables.

```console
$ helm repo add luizalabs http://helm.k8s.magazineluiza.com
$ helm install luizalabs/teresa \
  --name my-release \
  --namespace my-teresa \
  --set rsa.private=$TERESA_RSA_PRIVATE \
  --set rsa.public=$TERESA_RSA_PUBLIC \
  --set aws.key.access=XXXXXXXXXXXX \
  --set aws.key.secret=XXXXXXXXXXXX \
  --set aws.s3.bucket=teresa
```
This deploy teresa to cluster with default configuration.
The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```console
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following tables lists the configurable parameters of the Teresa chart and their default values.

Parameter | Description | Default
--------- | ----------- | -------
`name` | Deploy name | `teresa`
`db.name` | Database name | `teresa.sqlite`
`db.hostname`| (Optional) Database hostname, if defined use mysql instead of sqlite| `""`
`db.username` | (Optional) Database username | `""`
`db.password` | (Optional) Database password | `""`
`storage.type` | Type of storage | `s3`
`aws.s3.force_path_style` | To force path style instead of subdomain-style | `false`
`aws.s3.bucket` | S3 bucket path | `""`
`aws.s3.endpoint` | (Optional) AWS Endpoint | `""`
`aws.region` | AWS Region | `us-east-1`
`aws.key.access` | AWS Access Key | `""`
`aws.key.secret` | AWS Secret Key | `""`
`rsa.public` | RSA Public Key | `""`
`rsa.private` | RSA Private Key | `""`
`tls.crt` | (Optional) TLS Certificate | `""`
`tls.key` | (Optional) TLS Certificate Key | `""`
`docker.registry` | Docker Registry | `luizalabs` 
`docker.image` | Docker Image | `teresa`
`docker.tag` | Docker Tag | `0.5.0`
`build.limits.cpu` | CPU limit used by build POD  | `500m`
`build.limits.memory` | Memory limit used by build POD | `1024Mi`
`debug` | If true, print the stack trace on every panic/recover. | `false`

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```console
$ helm install luizalabs/teresa --name my-release \
    --set aws.region=us-east-2
```

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
$ helm install luizalabs/teresa --name my-release -f values.yaml
```
