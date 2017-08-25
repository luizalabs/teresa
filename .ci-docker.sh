#!/bin/bash

set -e

if [ -z "$TRAVIS_TAG" ]; then
    echo "skip docker imagem generation (no tag detected)"
    exit 0
fi

echo "Generate and upload the latest TERESA-SERVER image to Docker HUB"
make docker-login docker-build docker-push

exit 0;

