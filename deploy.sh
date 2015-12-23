#!/usr/bin/env bash
VERSION=0.1.0

export GITHUB_REPO=oracle-java-launcher
export GITHUB_USER=echocat

function checkExit {
	if [ $1 -ne 0 ]; then
		echo "Unexpected exit of child process. Got: ${1}" 1>&2
		exit $1
	fi
}

function uploadForPlattform {
	if [ -z "$2" ]; then
		echo "Parameter missing" 1>&2
		exit 1
	fi
	"${BUILD_DIR}/github-release" upload --tag "v${VERSION}" -n "oracle-java-launcher-${1}-${2}${3}" -f "${BUILD_DIR}/oracle-java-launcher-${1}-${2}${3}"
	checkExit $?
}


PLAIN_BASE="`dirname \"${0}\"`"
BASE="`readlink -f \"${PLAIN_BASE}\"`"
BUILD_DIR="${BASE}/target"

## Download release tool
curl -sSL "https://github.com/aktau/github-release/releases/download/v0.6.2/linux-amd64-github-release.tar.bz2" | tar -C "${BUILD_DIR}" -xjO > "${BUILD_DIR}/github-release"
checkExit $?
chmod +x "${BUILD_DIR}/github-release"
checkExit $?

## Remove eventually already existing version
"${BUILD_DIR}/github-release" delete --tag "v${VERSION}"

## Release the version
"${BUILD_DIR}/github-release" release --tag "v${VERSION}"
checkExit $?

## Upload artifacts
uploadForPlattform linux amd64
uploadForPlattform linux 386
uploadForPlattform windows amd64 .exe
uploadForPlattform windows 386 .exe
uploadForPlattform darwin amd64
uploadForPlattform darwin 386

