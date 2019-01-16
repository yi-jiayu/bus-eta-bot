#!/usr/bin/env bash

set -e

env="${1:-staging}"

# use service account credentials if we are on Travis
if [[ "$TRAVIS" = "true" ]]; then
    gcloud auth activate-service-account --key-file travis-ci-service-account.json
else
    tmpdir=$(mktemp -d)
    pushd ${tmpdir}
    cp -a $OLDPWD/ ./
    cleanup() {
        popd
        rm -rf ${tmpdir}
    }
    trap cleanup EXIT
fi

# embed current version into source
tag=$(git describe --tags)
date=$(date +%Y-%m-%d)
sed "s/VERSION/$tag/" < version.go > version.go~ && mv version.go~ version.go
sed "s/VERSION/$tag/; s/DATE/$date/" < index.html > index.html~ && mv index.html~ index.html

# version cannot contain dots
tag_escaped=$(sed "s/\./-/g" <<< ${tag})

# if environment is prod and tag is a release candidate, don't promote
if [[ ${env} = "prod" && "$tag" == *"rc"* ]]; then
    deploy_flags="--no-promote"
fi

gcloud --quiet app --project bus-eta-bot deploy ${deploy_flags} ${env}.app.yaml --version ${tag_escaped}
