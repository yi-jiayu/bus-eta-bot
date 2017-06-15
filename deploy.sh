#!/usr/bin/env bash

GCLOUD_PATH="${1:-gcloud}"
ENV="${2:-staging}"

$GCLOUD_PATH auth activate service account --key-file travis-ci-service-account.json
$GCLOUD_PATH --quiet app --project bus-eta-bot deploy ${ENV}.app.yaml --version $(sed "s/\./-/g" <<< $(git describe --tags))
