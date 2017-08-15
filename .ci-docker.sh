#!/bin/bash

set -e

if [ -z "$TRAVIS_TAG" ]; then
    echo "skip docker imagem generation (no tag detected)"
    exit 0
fi

echo "Generate and upload the latest TERESA-SERVER image to Docker HUB"
docker login -u "$DOCKER_USER" -p "$DOCKER_PASS"
docker build -t "$DOCKER_REGISTRY"/"$DOCKER_IMAGE":"$TRAVIS_TAG" .
docker push "$DOCKER_REGISTRY"/"$DOCKER_IMAGE":"$TRAVIS_TAG"

exit 0;

