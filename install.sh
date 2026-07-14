#!/usr/bin/env bash
# rescheck(resource_checker)와 instsched(instance_scheduler)를 각각 독립 모듈로
# 빌드해 /usr/local/bin (PATH)에 설치한다. 두 모듈은 서로 다른 go.mod를 갖는
# 완전히 독립된 Go 모듈이므로 각자의 디렉토리에서 개별적으로 빌드한다.
set -euo pipefail

VERSION="${VERSION:-v1.0.0}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if ! command -v go >/dev/null 2>&1; then
  echo "error: go가 설치되어 있지 않습니다. https://go.dev/dl/ 참고" >&2
  exit 1
fi

BUILD_DIR="$(mktemp -d)"
trap 'rm -rf "$BUILD_DIR"' EXIT

if [ -w "$INSTALL_DIR" ]; then
  SUDO=""
else
  SUDO="sudo"
fi

# build_binary <module_dir> <binary_name> <version_ldflags_path>
build_binary() {
  local module_dir="$1" binary_name="$2" version_pkg="$3"

  echo "==> ${binary_name} ${VERSION} 빌드 중 (${module_dir})..."
  (
    cd "${SCRIPT_DIR}/${module_dir}"
    go build -ldflags "-X '${version_pkg}.version=${VERSION}'" \
      -o "${BUILD_DIR}/${binary_name}" .
  )

  echo "==> ${INSTALL_DIR}/${binary_name} 설치 중 (${SUDO:-no sudo})..."
  ${SUDO} mkdir -p "$INSTALL_DIR"
  ${SUDO} install -m 755 "${BUILD_DIR}/${binary_name}" "${INSTALL_DIR}/${binary_name}"
}

build_binary "resource_checker" "rescheck" "github.com/opp-13/nhn_iac/resource_checker/cli"
build_binary "instance_scheduler" "instsched" "github.com/opp-13/nhn_iac/instance_scheduler/cli"

echo "==> 설치 완료: ${INSTALL_DIR}/rescheck, ${INSTALL_DIR}/instsched"

case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo
    echo "경고: ${INSTALL_DIR}가 PATH에 없습니다. 쉘 설정 파일(~/.bashrc, ~/.zshrc 등)에 다음을 추가하세요:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac

"${INSTALL_DIR}/rescheck" --version
"${INSTALL_DIR}/instsched" --version
