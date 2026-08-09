#!/bin/sh
set -eu

binary=${1:?usage: install.sh /path/to/webguard-server-agent}
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
test -x "$binary"
id -u webguard-server-agent >/dev/null 2>&1 || useradd --system --home /nonexistent --shell /usr/sbin/nologin webguard-server-agent
install -D -m 0755 "$binary" /usr/bin/webguard-server-agent
install -d -m 0750 -o webguard-server-agent -g webguard-server-agent /etc/webguard-server-agent /var/lib/webguard-server-agent/queue
if [ ! -f /etc/webguard-server-agent/config.json ]; then
  install -m 0600 -o webguard-server-agent -g webguard-server-agent "$script_dir/../config.example.json" /etc/webguard-server-agent/config.json
fi
install -D -m 0644 "$script_dir/../packaging/webguard-server-agent.service" /etc/systemd/system/webguard-server-agent.service
systemctl daemon-reload
echo "Edit /etc/webguard-server-agent/config.json, then run: systemctl enable --now webguard-server-agent"
