#!/usr/bin/env bash

set -e

# source .telegramrc file in working directory if it exists
[[ -f ".telegramrc" ]] && source .telegramrc

if [[ -z "$TELEGRAM_BOT_TOKEN" ]]; then
    echo TELEGRAM_BOT_TOKEN not set >&2
    exit 1
fi

echo "getting ngrok public url..."
public_url=$(curl localhost:4040/api/tunnels/command_line  2>/dev/null | jq -r '.public_url')
if [[ -z "$public_url" || "$public_url" == "null" ]]; then
    echo "failed to get public url. is ngrok running?"
    exit 1
fi
echo $public_url

webhook_url="$public_url/$TELEGRAM_BOT_TOKEN"
echo "setting webhook to $webhook_url..."
curl "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/setWebhook?url=$webhook_url" 2>/dev/null
echo
echo "starting dev server..."
dev_appserver.py  --enable_watching_go_path False --enable_host_checking=False web/dev.app.yaml
