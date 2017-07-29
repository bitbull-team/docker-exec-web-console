#!/bin/sh

set -e

socat TCP-LISTEN:2735,reuseaddr,fork UNIX-CONNECT:/var/run/docker.sock &

exec "$@"
