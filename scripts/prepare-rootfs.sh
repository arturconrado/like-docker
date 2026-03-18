#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET_DIR="${1:-${ROOT_DIR}/examples/rootfs/demo}"

if [[ "$(uname -s)" != "Linux" ]]; then
  echo "[minidock] prepare-rootfs é suportado apenas em Linux."
  echo "[minidock] o MVP continuará funcional via processo-local/demo."
  exit 0
fi

mkdir -p "${TARGET_DIR}"/{bin,usr/bin,lib,lib64,proc,tmp,etc,dev}

copy_binary_with_libs() {
  local bin_path="$1"

  if [[ ! -x "${bin_path}" ]]; then
    return 1
  fi

  local destination="${TARGET_DIR}${bin_path}"
  mkdir -p "$(dirname "${destination}")"
  cp -L "${bin_path}" "${destination}"

  while IFS= read -r lib; do
    [[ -z "${lib}" ]] && continue
    local lib_dest="${TARGET_DIR}${lib}"
    mkdir -p "$(dirname "${lib_dest}")"
    cp -L "${lib}" "${lib_dest}"
  done < <(ldd "${bin_path}" | awk '{for (i=1; i<=NF; i++) if ($i ~ /^\//) print $i}' | sort -u)

  return 0
}

copy_or_warn() {
  local name="$1"
  local found
  found="$(command -v "${name}" || true)"
  if [[ -z "${found}" ]]; then
    echo "[minidock] aviso: binário ${name} não encontrado no host."
    return
  fi
  if ! copy_binary_with_libs "${found}"; then
    echo "[minidock] aviso: não foi possível copiar ${name}."
  fi
}

copy_or_warn sh
copy_or_warn ls
copy_or_warn echo
copy_or_warn sleep
copy_or_warn uname
copy_or_warn id
copy_or_warn ps
copy_or_warn hostname
copy_or_warn pwd
copy_or_warn postgres
copy_or_warn initdb
copy_or_warn pg_ctl

if [[ ! -x "${TARGET_DIR}/bin/sh" ]]; then
  echo "[minidock] erro: rootfs sem /bin/sh executável após preparação."
  echo "[minidock] valide o host Linux e os binários locais antes de tentar container-linux."
  exit 1
fi

echo "mdk-demo" > "${TARGET_DIR}/etc/hostname"
chmod 1777 "${TARGET_DIR}/tmp" || true

echo "[minidock] rootfs de demonstração preparado em: ${TARGET_DIR}"
echo "[minidock] para usar: export MINIDOCK_CONTAINER_ROOTFS=${TARGET_DIR}"
echo "[minidock] execute API com privilégios root para container-linux real (ex.: sudo make api)."
