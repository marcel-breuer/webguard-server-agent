#!/bin/sh
set -eu

repository_directory=${1:?usage: test-apt-install.sh REPOSITORY_DIRECTORY KEYRING VERSION}
keyring=${2:?usage: test-apt-install.sh REPOSITORY_DIRECTORY KEYRING VERSION}
version=${3:?usage: test-apt-install.sh REPOSITORY_DIRECTORY KEYRING VERSION}

for image in debian:12-slim ubuntu:24.04; do
  docker run --rm \
    -v "$(cd "$repository_directory" && pwd):/repository:ro" \
    -v "$(cd "$(dirname "$keyring")" && pwd)/$(basename "$keyring"):/webguard-server-agent-archive-keyring.asc:ro" \
    "$image" \
    sh -ceu '
      export DEBIAN_FRONTEND=noninteractive
      apt-get update
      apt-get install -y --no-install-recommends ca-certificates gpg
      install -d -m 0755 /etc/apt/keyrings
      install -m 0644 /webguard-server-agent-archive-keyring.asc /etc/apt/keyrings/webguard-server-agent.asc
      echo "deb [signed-by=/etc/apt/keyrings/webguard-server-agent.asc] file:/repository stable main" > /etc/apt/sources.list.d/webguard-server-agent.list
      apt-get update
      apt-get install -y webguard-server-agent
      test -f /etc/webguard-server-agent/config.json
      test "$(webguard-server-agent --version)" = "$0"
    ' "$version"
done
