#!/usr/bin/env sh
set -e

HOST="$1"
shift
PORT="$1"
shift

while ! nc -z "$HOST" "$PORT"; do
  echo "waiting for $HOST:$PORT"
  sleep 1
done

exec "$@"
