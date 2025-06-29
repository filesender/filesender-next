#!/bin/sh

REGISTRY="codeberg.org"
IMAGE_NAME="codeberg.org/hattorius/filesender:latest"

# delete image if already exists
podman image rm "${IMAGE_NAME}" --force

# pull / update images needed
podman pull "golang:latest"
podman pull "alpine:latest"

# build image
podman build -t "${IMAGE_NAME}" .

if ! podman login --get-login "${REGISTRY}"; then
    # login to image registry
    podman login "${REGISTRY}"
fi

podman push "${IMAGE_NAME}"
