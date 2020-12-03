#!/usr/bin/env bash

initArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv5*) ARCH="armv5";;
    armv6*) ARCH="armv6";;
    armv7*) ARCH="arm";;
    aarch64) ARCH="arm64";;
    x86) ARCH="386";;
    x86_64) ARCH="amd64";;
    i686) ARCH="386";;
    i386) ARCH="386";;
  esac
}

initOS() {
  OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')
}

isSupported() {
  local supported="darwin-amd64\ndarwin-arm64\nlinux-amd64\nlinux-arm\nlinux-arm64\nwindows-amd64\nwindows-arm"
  if ! echo "${supported}" | grep -q "${OS}-${ARCH}"; then
    echo "No prebuilt binary for ${OS}-${ARCH}."
    echo "To build from source, go to https://github.com/hazelcast/hazelcast-cloud-cli"
    exit 1
  fi

  if ! type "curl" > /dev/null && ! type "wget" > /dev/null; then
    echo "Either curl or wget is required"
    exit 1
  fi
}

download() {
  HZCLOUD_DIST="hzcloud-$OS-$ARCH"
  DOWNLOAD_URL="https://github.com/hazelcast/hazelcast-cloud-cli/releases/latest/download/$HZCLOUD_DIST"
  HZCLOUD_TMP_ROOT="$(mktemp -dt hzcloud-installer-XXXXXX)"
  HZCLOUD_TMP_FILE="$HZCLOUD_TMP_ROOT/$HZCLOUD_DIST"
  echo "Downloading $DOWNLOAD_URL"
  if type "curl" > /dev/null; then
    curl -SsL "$DOWNLOAD_URL" -o "$HZCLOUD_TMP_FILE"
  elif type "wget" > /dev/null; then
    wget -q -O "$HZCLOUD_TMP_FILE" "$DOWNLOAD_URL"
  fi
}

runAsRoot() {
  local CMD="$*"
  if [ $EUID -ne 0 ]; then
    CMD="sudo $CMD"
  fi
  $CMD
}

install() {
  INSTALL_DIR="/usr/local/bin"
  INSTALL_NAME="hzcloud"
  runAsRoot mv "$HZCLOUD_TMP_FILE" "$INSTALL_DIR/$INSTALL_NAME"
  runAsRoot chmod +x "$INSTALL_DIR/$INSTALL_NAME"
  echo "Installation finished. Run hzcloud to start!"
}

initArch
initOS
isSupported
download
install