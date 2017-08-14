#!/bin/bash

set -e

function download {
	curl https://storage.googleapis.com/kubernetes-helm/helm-v2.5.1-linux-amd64.tar.gz --output /tmp/helm.tar.gz --silent

	tar -x -C /tmp -f /tmp/helm.tar.gz linux-amd64/helm
}

function lint {
	/tmp/linux-amd64/helm lint helm/chart/teresa
}

case $1 in
	"download" )
		download ;;
	"lint" )
		lint ;;
esac
