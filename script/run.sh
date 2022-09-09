#!/bin/sh

set -eu

cd "$(dirname "${0}")/../"

docker run -e SLACK_BOT_TOKEN="${SLACK_BOT_TOKEN}" -e RECYCLE_LIST="${RECYCLE_LIST}" bidonbot
