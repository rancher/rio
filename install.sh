#!/usr/bin/env bash

APP_NAME="rio"
REPO_URL="https://github.com/rancher/rio"

: ${USE_SUDO:="true"}
: ${RIO_INSTALL_DIR:="/usr/local/bin"}

# initArch discovers the architecture for this system.
initArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv*) ARCH="arm";;
    aarch64) ARCH="arm64";;
    x86) ARCH="386";;
    x86_64) ARCH="amd64";;
    i686) ARCH="386";;
    i386) ARCH="386";;
  esac
}

# initOS discovers the operating system for this system.
initOS() {
  OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

  case "$OS" in
    # Minimalist GNU for Windows
    mingw*) OS='windows';;
  esac
}

# runs the given command as root (detects if we are root already)
runAsRoot() {
  local CMD="$*"

  if [ $EUID -ne 0 -a $USE_SUDO = "true" ]; then
    CMD="sudo $CMD"
  fi

  $CMD
}

# verifySupported checks that the os/arch combination is supported for
# binary builds.
verifySupported() {
  local supported="darwin-amd64\nlinux-amd64\nlinux-arm\nlinux-arm64\nwindows-amd64"
  if ! echo "${supported}" | grep -q "${OS}-${ARCH}"; then
    echo "No prebuilt binary for ${OS}-${ARCH}."
    echo "To build from source, go to $REPO_URL"
    exit 1
  fi

  if ! type "curl" > /dev/null && ! type "wget" > /dev/null; then
    echo "Either curl or wget is required"
    exit 1
  fi
}

# checkRioInstalledVersion checks which version of rio is installed and
# if it needs to be changed.
checkRioInstalledVersion() {
  if [[ -f "${RIO_INSTALL_DIR}/${APP_NAME}" ]]; then
    local version=$(rio --version | cut -d " " -f3)
    if [[ "$version" == "$TAG" ]]; then
      echo "rio ${version} is already ${DESIRED_VERSION:-latest}"
      return 0
    else
      echo "rio ${TAG} is available. Changing from version ${version}."
      return 1
    fi
  else
    return 1
  fi
}

# checkLatestVersion grabs the latest version string from the releases
checkLatestVersion() {
  local latest_release_url="$REPO_URL/releases/latest"
  if type "curl" > /dev/null; then
    TAG=$(curl -Ls -o /dev/null -w %{url_effective} $latest_release_url | grep -oE "[^/]+$" )
  elif type "wget" > /dev/null; then
    TAG=$(wget $latest_release_url --server-response -O /dev/null 2>&1 | awk '/^  Location: /{DEST=$2} END{ print DEST}' | grep -oE "[^/]+$")
  fi
}

# downloadFile downloads the latest binary package and also the checksum
# for that binary.
downloadFile() {
  RIO_DIST="rio-$OS-$ARCH"
  DOWNLOAD_URL="$REPO_URL/releases/download/$TAG/$RIO_DIST"
  RIO_TMP_ROOT="$(mktemp -dt rio-binary-XXXXXX)"
  RIO_TMP_FILE="$RIO_TMP_ROOT/$RIO_DIST"
  if type "curl" > /dev/null; then
    curl -SsL "$DOWNLOAD_URL" -o "$RIO_TMP_FILE"
  elif type "wget" > /dev/null; then
    wget -q -O "$RIO_TMP_FILE" "$DOWNLOAD_URL"
  fi
}

# installFile verifies the SHA256 for the file, then unpacks and
# installs it.
installFile() {
  echo "Preparing to install $APP_NAME into ${RIO_INSTALL_DIR}"
  runAsRoot chmod +x "$RIO_TMP_FILE"
  runAsRoot cp "$RIO_TMP_FILE" "$RIO_INSTALL_DIR/$APP_NAME"
  echo "$APP_NAME installed into $RIO_INSTALL_DIR/$APP_NAME"
}

# fail_trap is executed if an error occurs.
fail_trap() {
  result=$?
  if [ "$result" != "0" ]; then
    if [[ -n "$INPUT_ARGUMENTS" ]]; then
      echo "Failed to install $APP_NAME with the arguments provided: $INPUT_ARGUMENTS"
      help
    else
      echo "Failed to install $APP_NAME"
    fi
    echo -e "\tFor support, go to $REPO_URL."
  fi
  cleanup
  exit $result
}

# testVersion tests the installed client to make sure it is working.
testVersion() {
  set +e
  RIO="$(which $APP_NAME)"
  if [ "$?" = "1" ]; then
    echo "$APP_NAME not found. Is $RIO_INSTALL_DIR on your "'$PATH?'
    exit 1
  fi
  set -e
  echo "Run '$APP_NAME --help' to see what you can do with it."
}

# help provides possible cli installation arguments
help () {
  echo "Accepted cli arguments are:"
  echo -e "\t[--help|-h ] ->> prints this help"
  echo -e "\t[--no-sudo]  ->> install without sudo"
}

# cleanup temporary files
cleanup() {
  if [[ -d "${RIO_TMP_ROOT:-}" ]]; then
    rm -rf "$RIO_TMP_ROOT"
  fi
}

# Execution

#Stop execution on any error
trap "fail_trap" EXIT
set -e

# Parsing input arguments (if any)
export INPUT_ARGUMENTS="${@}"
set -u
while [[ $# -gt 0 ]]; do
  case $1 in
    '--no-sudo')
       USE_SUDO="false"
       ;;
    '--help'|-h)
       help
       exit 0
       ;;
    *) exit 1
       ;;
  esac
  shift
done
set +u

initArch
initOS
verifySupported
checkLatestVersion
if ! checkRioInstalledVersion; then
  downloadFile
  installFile
fi
testVersion
cleanup
