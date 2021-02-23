#!/bin/bash

set -e

function setup {
	curl https://get.helm.sh/helm-v2.17.0-linux-amd64.tar.gz --output /tmp/helm.tar.gz --silent

	tar -x -C /tmp -f /tmp/helm.tar.gz linux-amd64/helm
	/tmp/linux-amd64/helm init --client-only

	if [[ -n $TRAVIS_TAG ]]; then
		pip install --user awscli
	fi
}

function lint {
	/tmp/linux-amd64/helm lint helm/chart/teresa
}

function dependency-update {
	/tmp/linux-amd64/helm repo update
	/tmp/linux-amd64/helm dependency update helm/chart/teresa
}

function deploy {
	/tmp/linux-amd64/helm package helm/chart/teresa
	mkdir /tmp/repo
	mv teresa-*.tgz /tmp/repo
	/tmp/linux-amd64/helm repo index /tmp/repo --url http://helm.k8s.magazineluiza.com
	aws s3 sync /tmp/repo s3://helm.k8s.magazineluiza.com --delete
}

case $1 in
	"setup" )
		setup ;;
	"lint" )
		lint ;;
	"dependency-update" )
		dependency-update ;;
	"deploy")
		if [[ -z "$TRAVIS_TAG" ]]; then
			echo "skip helm repo update (no tag detected)"
			exit 0
		fi
		deploy
		;;
esac
