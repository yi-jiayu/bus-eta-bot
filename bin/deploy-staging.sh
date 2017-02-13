#!/usr/bin/env bash

set -e

echo $GCLOUD_SERVICE_KEY | base64 --decode > ${HOME}/gcloud-service-key.json
sudo /opt/google-cloud-sdk/bin/gcloud --quiet components update
sudo /opt/google-cloud-sdk/bin/gcloud --quiet components install alpha
sudo /opt/google-cloud-sdk/bin/gcloud auth activate-service-account --key-file ${HOME}/gcloud-service-key.json
sudo /opt/google-cloud-sdk/bin/gcloud config set project $GCLOUD_PROJECT
npm run babel
echo $STAGING_BOT_ENV | base64 --decode > .env.json
sudo /opt/google-cloud-sdk/bin/gcloud alpha functions deploy bus-eta-bot-staging --stage-bucket bus-eta-bot-src --entry-point main --memory 128MB --timeout 60 --trigger-http
