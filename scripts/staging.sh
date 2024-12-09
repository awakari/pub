#!/bin/bash

export SLUG=ghcr.io/awakari/pub
export VERSION=latest
docker tag awakari/pub "${SLUG}":"${VERSION}"
docker push "${SLUG}":"${VERSION}"
