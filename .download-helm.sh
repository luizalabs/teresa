#!/usr/bin/env bash

set -e

wget https://storage.googleapis.com/kubernetes-helm/helm-v2.5.1-linux-amd64.tar.gz /tmp/helm.tar.gz
tar -xcf /tmp/helm.tar.gz
export PATH=$PATH:$PWD/linux-amd64
