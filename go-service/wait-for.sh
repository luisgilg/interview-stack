#!/usr/bin/env sh
set -e

host="$1"
shift
port="$1"
shift

while ! nc -z "$host" "$port"; do
  echo "waiting for $host:$port"
  sleep 1
done

exec "$@"
