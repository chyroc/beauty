#!/usr/bin/env bash

set -e
set -x
set -o pipefail

go install github.com/chyroc/action.sh/commiter@v0.4.0 && mv `which commiter` /tmp/commiter
go build -o /tmp/fetch ./src/command/fetch.go
go build -o /tmp/send_msg ./src/command/send_msg.go
