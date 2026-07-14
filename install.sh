#!/usr/bin/env bash
# rescheck를 빌드해 /usr/local/bin (PATH)에 설치한다.
set -euo pipefail

BINARY_NAME="rescheck"
VERSION="${VERSION:-v1.0.0}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
MODULE="github.com/opp-13/nhn_iac"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if ! command -v go >/dev/null 2>&1; then
  echo "error: go가 설치되어 있지 않습니다. https://go.dev/dl/ 참고" >&2
  exit 1
fi

echo "==> ${BINARY_NAME} ${VERSION} 빌드 중..."
BUILD_DIR="$(mktemp -d)"
trap 'rm -rf "$BUILD_DIR"' EXIT
go build -ldflags "-X '${MODULE}/resource_checker/cli.version=${VERSION}'" \
  -o "${BUILD_DIR}/${BINARY_NAME}" .

if [ -w "$INSTALL_DIR" ]; then
  SUDO=""
else
  SUDO="sudo"
fi

echo "==> ${INSTALL_DIR}/${BINARY_NAME} 설치 중 (${SUDO:-no sudo})..."
${SUDO} mkdir -p "$INSTALL_DIR"
${SUDO} install -m 755 "${BUILD_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"

echo "==> 설치 완료: $(command -v "$BINARY_NAME" 2>/dev/null || echo "${INSTALL_DIR}/${BINARY_NAME}")"

case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo
    echo "경고: ${INSTALL_DIR}가 PATH에 없습니다. 쉘 설정 파일(~/.bashrc, ~/.zshrc 등)에 다음을 추가하세요:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac

"${INSTALL_DIR}/${BINARY_NAME}" --version
