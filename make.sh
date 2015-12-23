#!/usr/bin/env bash
set -e

function checkExit {
	if [ $1 -ne 0 ]; then
		echo "Unexpected exit of child process. Got: ${1}" 1>&2
		exit $1
	fi
}

function buildForPlattform {
	if [ -z "$2" ]; then
		echo "Parameter missing" 1>&2
		exit 1
	fi
	go-crosscompile-build "$1/$2"
	checkExit $?

	GOOS="$1" GOARCH="$2" "${GOROOT}/bin/go" build -o "${BUILD_DIR}/oracle-java-launcher-${1}-${2}${3}" "${BASE}/main.go"
	checkExit $?
}

GO_VERSION=1.5.2
PLAIN_BASE="`dirname \"${0}\"`"
BASE="`readlink -f \"${PLAIN_BASE}\"`"

export BUILD_DIR="${BASE}/target"
export GOROOT="${BUILD_DIR}/go"
export GOROOT_BOOTSTRAP="${BUILD_DIR}/go-bootstrap"
export GOPATH="${BUILD_DIR}/gohome"

## Clean build directory
rm -rf "${BUILD_DIR}"
checkExit $?

## Create build directory
mkdir -p "${BUILD_DIR}"
checkExit $?

## Download go bootstrap and extract
curl -sSL "https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz" | tar -C "${BUILD_DIR}" -xz
checkExit $?
mv "${BUILD_DIR}/go" "${BUILD_DIR}/go-bootstrap"
checkExit $?

## Download go sources and extract
curl -sSL "https://golang.org/dl/go${GO_VERSION}.src.tar.gz" | tar -C "${BUILD_DIR}" -xz
checkExit $?

## Download go crosscompile tools
curl -sSL "https://raw.githubusercontent.com/davecheney/golang-crosscompile/master/crosscompile.bash" > "${BUILD_DIR}/crosscompile.bash"
checkExit $?
source "${BUILD_DIR}/crosscompile.bash"

## Build go
cd "${GOROOT}/src"
checkExit $?
./make.bash
checkExit $?

cd "${BASE}"
checkExit $?

## Start build of app
buildForPlattform linux amd64
buildForPlattform linux 386
buildForPlattform windows amd64 .exe
buildForPlattform windows 386 .exe
buildForPlattform darwin amd64
buildForPlattform darwin 386

