#!/bin/sh
set -eu

version=${1:?usage: build-deb.sh VERSION ARCHITECTURE [OUTPUT_DIRECTORY]}
architecture=${2:?usage: build-deb.sh VERSION ARCHITECTURE [OUTPUT_DIRECTORY]}
output_directory=${3:-dist/deb}

version=${version#v}
case "$version" in
  [0-9]*.[0-9]*.[0-9]*) ;;
  *) echo "VERSION must be a semantic version" >&2; exit 1 ;;
esac

case "$architecture" in
  amd64) go_architecture=amd64 ;;
  arm64) go_architecture=arm64 ;;
  *) echo "unsupported Debian architecture: $architecture" >&2; exit 1 ;;
esac

package_root=$(mktemp -d)
trap 'rm -rf "$package_root"' EXIT HUP INT TERM

mkdir -p "$package_root/DEBIAN" "$package_root/usr/bin" "$package_root/usr/share/webguard-server-agent" "$package_root/lib/systemd/system" "$output_directory"
sed -e "s/@VERSION@/$version/" -e "s/@ARCH@/$architecture/" packaging/debian/control.in > "$package_root/DEBIAN/control"
install -m 0755 packaging/debian/preinst packaging/debian/postinst packaging/debian/prerm packaging/debian/postrm "$package_root/DEBIAN/"
install -m 0644 config.example.json "$package_root/usr/share/webguard-server-agent/config.example.json"
install -m 0644 packaging/webguard-server-agent.service "$package_root/lib/systemd/system/webguard-server-agent.service"
CGO_ENABLED=0 GOOS=linux GOARCH="$go_architecture" go build -trimpath -ldflags "-s -w -X main.version=v$version" -o "$package_root/usr/bin/webguard-server-agent" ./cmd/webguard-server-agent
dpkg-deb --root-owner-group --build "$package_root" "$output_directory/webguard-server-agent_${version}_${architecture}.deb"
