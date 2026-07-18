#!/usr/bin/env bash
# Build a menu-bar .app so the binary can run outside Terminal
# (Accessibility / Microphone attach to the .app, not to Terminal).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP_NAME="ShutUpAndType"
APP_DIR="${1:-$HOME/Applications}/${APP_NAME}.app"
BIN_SRC="$ROOT/shutupandtype-x11"
MACOS_DIR="$APP_DIR/Contents/MacOS"
RES_DIR="$APP_DIR/Contents/Resources"

cd "$ROOT"
CGO_ENABLED=1 go build -o "$BIN_SRC" .

rm -rf "$APP_DIR"
mkdir -p "$MACOS_DIR" "$RES_DIR"
cp "$BIN_SRC" "$MACOS_DIR/$APP_NAME"
chmod +x "$MACOS_DIR/$APP_NAME"

if [[ -f "$ROOT/icon.png" ]]; then
  # Optional: best-effort icns via sips+iconutil when available
  ICONSET="$(mktemp -d)/App.iconset"
  mkdir -p "$ICONSET"
  for s in 16 32 128 256 512; do
    sips -z "$s" "$s" "$ROOT/icon.png" --out "$ICONSET/icon_${s}x${s}.png" >/dev/null 2>&1 || true
    sips -z $((s*2)) $((s*2)) "$ROOT/icon.png" --out "$ICONSET/icon_${s}x${s}@2x.png" >/dev/null 2>&1 || true
  done
  if iconutil -c icns "$ICONSET" -o "$RES_DIR/AppIcon.icns" 2>/dev/null; then
    ICON_KEY=$'\n  <key>CFBundleIconFile</key>\n  <string>AppIcon</string>'
  else
    ICON_KEY=""
  fi
  rm -rf "$(dirname "$ICONSET")"
else
  ICON_KEY=""
fi

cat > "$APP_DIR/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key>
  <string>${APP_NAME}</string>
  <key>CFBundleIdentifier</key>
  <string>com.harnyk.shutupandtype</string>
  <key>CFBundleName</key>
  <string>${APP_NAME}</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>CFBundleShortVersionString</key>
  <string>0.1.0</string>
  <key>CFBundleVersion</key>
  <string>1</string>
  <key>LSMinimumSystemVersion</key>
  <string>13.0</string>
  <key>LSUIElement</key>
  <true/>
  <key>NSMicrophoneUsageDescription</key>
  <string>Needed to record speech for transcription.</string>${ICON_KEY}
</dict>
</plist>
EOF

# Clear quarantine so first launch is less painful
xattr -dr com.apple.quarantine "$APP_DIR" 2>/dev/null || true

# Ad-hoc sign with a stable identifier so Accessibility TCC survives rebuilds better.
codesign --force --deep --sign - --identifier com.harnyk.shutupandtype "$APP_DIR" 2>/dev/null || true

echo "Installed: $APP_DIR"
echo "Open with: open \"$APP_DIR\""
echo "Then grant Accessibility + Microphone to ${APP_NAME} in System Settings."
echo "If it won't start after a rebuild: remove ${APP_NAME} from Accessibility, then add it again."

