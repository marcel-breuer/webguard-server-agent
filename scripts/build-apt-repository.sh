#!/bin/sh
set -eu

repository_directory=${1:?usage: build-apt-repository.sh REPOSITORY_DIRECTORY PACKAGE_DIRECTORY}
package_directory=${2:?usage: build-apt-repository.sh REPOSITORY_DIRECTORY PACKAGE_DIRECTORY}

: "${GNUPGHOME:?GNUPGHOME must contain the imported repository signing key}"
: "${GPG_KEY_ID:?GPG_KEY_ID must identify the repository signing key}"
: "${APT_REPOSITORY_SIGNING_PASSPHRASE:?APT_REPOSITORY_SIGNING_PASSPHRASE is required}"

mkdir -p "$repository_directory/pool/main/w/webguard-server-agent"
touch "$repository_directory/.nojekyll"
find "$package_directory" -maxdepth 1 -type f -name 'webguard-server-agent_*.deb' -exec cp {} "$repository_directory/pool/main/w/webguard-server-agent/" \;

for architecture in amd64 arm64; do
  packages_directory="$repository_directory/dists/stable/main/binary-$architecture"
  mkdir -p "$packages_directory"
  (
    cd "$repository_directory"
    dpkg-scanpackages --arch "$architecture" pool /dev/null > "dists/stable/main/binary-$architecture/Packages"
  )
  gzip --no-name --best --stdout "$packages_directory/Packages" > "$packages_directory/Packages.gz"
done

(
  cd "$repository_directory"
  apt-ftparchive \
    -o APT::FTPArchive::Release::Origin="WebGuard" \
    -o APT::FTPArchive::Release::Label="WebGuard Server Agent" \
    -o APT::FTPArchive::Release::Suite="stable" \
    -o APT::FTPArchive::Release::Codename="stable" \
    -o APT::FTPArchive::Release::Architectures="amd64 arm64" \
    -o APT::FTPArchive::Release::Components="main" \
    release dists/stable > dists/stable/Release
)

gpg --batch --yes --pinentry-mode loopback --passphrase "$APT_REPOSITORY_SIGNING_PASSPHRASE" --local-user "$GPG_KEY_ID" --armor --detach-sign --output "$repository_directory/dists/stable/Release.gpg" "$repository_directory/dists/stable/Release"
gpg --batch --yes --pinentry-mode loopback --passphrase "$APT_REPOSITORY_SIGNING_PASSPHRASE" --local-user "$GPG_KEY_ID" --clearsign --output "$repository_directory/dists/stable/InRelease" "$repository_directory/dists/stable/Release"
