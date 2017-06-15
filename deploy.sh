#!/usr/bin/env bash

ENV="${1:-staging}"

gcloud auth activate-service-account --key-file travis-ci-service-account.json
gcloud --quiet app --project bus-eta-bot deploy ${ENV}.app.yaml --version $(sed "s/\./-/g" <<< $(git describe --tags))
