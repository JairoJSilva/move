#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
STAGE_DIR="${SCRIPT_DIR}/stage"

echo "=================================================="
echo "  Gerando Pacote Debian (.deb) do MoveOps"
echo "=================================================="

# Garante compilação atualizada do binário Linux
echo "==> Compilando binário Linux final..."
cd "${ROOT_DIR}"
CGO_ENABLED=0 go build -ldflags="-s -w" -o "${ROOT_DIR}/bin/moveops" cmd/moveops/main.go

# Limpa e prepara área de staging
echo "==> Montando estrutura limpa em staging..."
rm -rf "${STAGE_DIR}"
mkdir -p "${STAGE_DIR}/DEBIAN"
mkdir -p "${STAGE_DIR}/usr/local/bin"
mkdir -p "${STAGE_DIR}/usr/share/applications"
mkdir -p "${STAGE_DIR}/usr/share/icons/hicolor/scalable/apps"
mkdir -p "${STAGE_DIR}/lib/systemd/system"
mkdir -p "${ROOT_DIR}/dist"

# Copia metadados DEBIAN
cp "${SCRIPT_DIR}/DEBIAN/control" "${STAGE_DIR}/DEBIAN/control"
cp "${SCRIPT_DIR}/DEBIAN/postinst" "${STAGE_DIR}/DEBIAN/postinst"
cp "${SCRIPT_DIR}/DEBIAN/prerm" "${STAGE_DIR}/DEBIAN/prerm"
chmod 755 "${STAGE_DIR}/DEBIAN/postinst"
chmod 755 "${STAGE_DIR}/DEBIAN/prerm"
chmod 644 "${STAGE_DIR}/DEBIAN/control"

# Copia arquivos do sistema
cp "${ROOT_DIR}/bin/moveops" "${STAGE_DIR}/usr/local/bin/moveops"
chmod 755 "${STAGE_DIR}/usr/local/bin/moveops"

cp "${ROOT_DIR}/logo-n-fundo.png" "${STAGE_DIR}/usr/share/icons/hicolor/scalable/apps/moveops.png"
chmod 644 "${STAGE_DIR}/usr/share/icons/hicolor/scalable/apps/moveops.png"

cp "${SCRIPT_DIR}/moveops.desktop" "${STAGE_DIR}/usr/share/applications/moveops.desktop"
chmod 644 "${STAGE_DIR}/usr/share/applications/moveops.desktop"

cp "${SCRIPT_DIR}/moveops.service" "${STAGE_DIR}/lib/systemd/system/moveops.service"
chmod 644 "${STAGE_DIR}/lib/systemd/system/moveops.service"

OUTPUT_DEB="${ROOT_DIR}/dist/moveops_1.0.0_amd64.deb"

echo "==> Construindo pacote com dpkg-deb..."
dpkg-deb --build --root-owner-group "${STAGE_DIR}" "${OUTPUT_DEB}"

# Limpa diretório temporário de staging
rm -rf "${STAGE_DIR}"

echo "=================================================="
echo "  Pacote gerado com sucesso:"
echo "  ${OUTPUT_DEB}"
ls -lh "${OUTPUT_DEB}"
echo "=================================================="
