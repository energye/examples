#!/bin/bash
# Download Noto Sans CJK SC Regular font for embedding in the gpui library.
# This is the Google version of Adobe Source Han Sans SC (思源黑体).
# Download size: ~16 MB, covers Latin + CJK (Chinese, Japanese, Korean).

set -euo pipefail

FONT_DIR="$(cd "$(dirname "$0")/../render/font/fonts" && pwd)"
FONT_NAME="NotoSansCJK-Regular"
FONT_FILE="${FONT_DIR}/${FONT_NAME}.ttc"

# Official Noto CJK release URL (Google Fonts / Noto repository)
# This is the latest stable release of Noto Sans CJK Regular.
# It covers Latin + CJK (Chinese, Japanese, Korean).
URL="https://github.com/notofonts/noto-cjk/releases/download/Sans2.004/03_NotoSansCJK.zip"

mkdir -p "$FONT_DIR"

if [ -f "$FONT_FILE" ]; then
    echo "✓ Font already exists: $FONT_FILE ($(du -h "$FONT_FILE" | cut -f1))"
    exit 0
fi

echo "Downloading Noto Sans CJK SC Regular from Noto repository..."
echo "URL: $URL"

TMP_DIR=$(mktemp -d)
trap "rm -rf '$TMP_DIR'" EXIT

cd "$TMP_DIR"

if command -v curl &>/dev/null; then
    curl -L -o noto-cjk.zip "$URL"
elif command -v wget &>/dev/null; then
    wget -O noto-cjk.zip "$URL"
else
    echo "Error: Neither curl nor wget found. Please install one of them."
    exit 1
fi

# Extract the TTC file
unzip -j noto-cjk.zip "*.ttc" -d extracted/ 2>/dev/null || unzip -j noto-cjk.zip "NotoSansCJK-Regular.ttc" -d extracted/

# Find the font file
TTC_FILE=$(find extracted/ -name "NotoSansCJK-Regular.ttc" | head -1)
if [ -z "$TTC_FILE" ]; then
    echo "Error: Could not find NotoSansCJK-Regular.ttc in the downloaded archive."
    echo "Contents of extracted/:"
    ls -la extracted/
    exit 1
fi

cp "$TTC_FILE" "$FONT_FILE"
echo "✓ Font downloaded: $FONT_FILE ($(du -h "$FONT_FILE" | cut -f1))"
echo ""
echo "Font info:"
fc-scan "$FONT_FILE" 2>/dev/null | head -5 || otfinfo -i "$FONT_FILE" 2>/dev/null | head -5 || true