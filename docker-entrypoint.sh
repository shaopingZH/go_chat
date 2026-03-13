#!/bin/sh
set -eu

UPLOAD_DIR="${UPLOAD_DIR:-/app/data/uploads}"

mkdir -p "$UPLOAD_DIR"
chown -R appuser:appuser "$UPLOAD_DIR"

exec su-exec appuser /app/server
