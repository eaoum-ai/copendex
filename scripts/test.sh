#!/usr/bin/env sh
set -eu

GO="${GO:-go}"

exec "$GO" test ./...
