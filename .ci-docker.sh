#!/bin/bash

set -e

echo "Generate and upload the latest TERESA-SERVER image to Docker HUB"
docker login -e "$DOCKER_EMAIL" -u "$DOCKER_USER" -p "$DOCKER_PASS"
docker build -t "$DOCKER_REGISTRY"/"$DOCKER_IMAGE":"$TRAVIS_TAG" .
docker push "$DOCKER_REGISTRY"/"$DOCKER_IMAGE":"$TRAVIS_TAG"

exit 0;
