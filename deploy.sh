#!/usr/bin/env bash

ENV="${1:-staging}"

$HOME/google-cloud-sdk/bin/gcloud auth activate service account --key-file travis-ci-service-account.json
$HOME/google-cloud-sdk/bin/gcloud --quiet app --project bus-eta-bot deploy ${ENV}.app.yaml --version $(sed "s/\./-/g" <<< $(git describe --tags))
