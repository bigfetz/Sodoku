#!/bin/zsh
# deploy_iphone.sh — builds and installs Sudoku on connected iPhone
#
# ── CONFIGURE THESE FOR YOUR MACHINE ────────────────────────────────────────
# CERT:        Your Apple Developer signing identity.
#              Find it with: security find-identity -v -p codesigning
CERT="YOUR_SIGNING_IDENTITY"          # e.g. "Apple Development: you@example.com (XXXXXXXXXX)"
#
# PROFILE:     Full path to your provisioning profile (.mobileprovision).
#              Download from https://developer.apple.com/account/resources/profiles/
#              or find existing ones in:
#              ~/Library/Developer/Xcode/UserData/Provisioning Profiles/
PROFILE="$HOME/Library/Developer/Xcode/UserData/Provisioning Profiles/YOUR_PROFILE_UUID.mobileprovision"
#
# DEVICE_UDID: UDID of the iPhone to install on.
#              Find it with: xcrun devicectl list devices
#              or in Xcode → Window → Devices and Simulators
DEVICE_UDID="YOUR_DEVICE_UDID"
#
# BUNDLE_ID:   Must match the bundle ID in your provisioning profile.
BUNDLE_ID="YOUR_BUNDLE_ID"           # e.g. "com.yourname.sudoku"
#
# TEAM_ID:     Your 10-character Apple Developer Team ID.
#              Find it at https://developer.apple.com/account → Membership
TEAM_ID="YOUR_TEAM_ID"               # e.g. "AB12CD34EF"
# ────────────────────────────────────────────────────────────────────────────
set -e

APPDIR="$(dirname "$0")/Sudoku.app"

echo "==> Building for iOS simulator (generates .app structure)..."
cd "$(dirname "$0")"
$(go env GOPATH)/bin/fyne package --target iossimulator --app-id "$BUNDLE_ID"

echo "==> Cross-compiling for arm64 device..."
CGO_ENABLED=1 GOOS=ios GOARCH=arm64 \
  CC=$(xcrun --sdk iphoneos --find clang) \
  CGO_CFLAGS="-isysroot $(xcrun --sdk iphoneos --show-sdk-path) -arch arm64 -miphoneos-version-min=12.0" \
  CGO_LDFLAGS="-isysroot $(xcrun --sdk iphoneos --show-sdk-path) -arch arm64 -miphoneos-version-min=12.0" \
  go build -tags ios -o /tmp/sudoku_arm64 .

echo "==> Injecting arm64 binary and provisioning profile..."
cp /tmp/sudoku_arm64 "$APPDIR/Sudoku"
chmod +x "$APPDIR/Sudoku"
cp "$PROFILE" "$APPDIR/embedded.mobileprovision"
/usr/libexec/PlistBuddy -c "Set :CFBundleIdentifier $BUNDLE_ID" "$APPDIR/Info.plist"
/usr/libexec/PlistBuddy -c "Set :CFBundleExecutable Sudoku" "$APPDIR/Info.plist"

echo "==> Signing..."
cat > /tmp/entitlements.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>application-identifier</key>
    <string>${TEAM_ID}.${BUNDLE_ID}</string>
    <key>com.apple.developer.team-identifier</key>
    <string>${TEAM_ID}</string>
    <key>get-task-allow</key>
    <true/>
</dict>
</plist>
EOF
codesign --force --sign "$CERT" --entitlements /tmp/entitlements.plist "$APPDIR"

echo "==> Installing on iPhone..."
xcrun devicectl device install app --device "$DEVICE_UDID" "$APPDIR"

echo ""
echo "✅ Done! Find 'Sudoku' on your iPhone home screen."
