#!/usr/bin/env zsh
# deploy_android.sh — Build and install the Sudoku app on a connected Android device.
# Enable USB debugging on the device: Settings → Developer options → USB debugging.
#
# ── CONFIGURE THESE FOR YOUR MACHINE ────────────────────────────────────────
# JAVA_HOME:        Path to the JDK bundled with Android Studio.
#                   Android Studio typically ships its own JBR; adjust the
#                   version path if yours differs.
#                   Find it: /usr/libexec/java_home  or check Android Studio
#                   → Settings → Build Tools → Gradle → Gradle JDK
export JAVA_HOME="/Applications/Android Studio.app/Contents/jbr/Contents/Home" # adjust JBR version path if needed
#
# ANDROID_NDK_HOME: Path to the NDK version you have installed.
#                   Check installed versions: ls $ANDROID_HOME/ndk/
#                   Install via Android Studio → SDK Manager → SDK Tools → NDK
export ANDROID_NDK_HOME="$ANDROID_HOME/ndk/YOUR_NDK_VERSION"   # e.g. 27.2.12479018
#
# BUILD_TOOLS_VER:  Version of Android build-tools you have installed.
#                   Check installed versions: ls $ANDROID_HOME/build-tools/
BUILD_TOOLS_VER="YOUR_BUILD_TOOLS_VERSION"                      # e.g. 35.0.0
#
# BUNDLE_ID:        Your Android app ID (reverse-domain format).
BUNDLE_ID="YOUR_BUNDLE_ID"                                      # e.g. com.yourname.sudoku
# ────────────────────────────────────────────────────────────────────────────

set -e

export PATH="$JAVA_HOME/bin:$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/build-tools/$BUILD_TOOLS_VER:$HOME/go/bin:$PATH"

echo "==> Building Android APK..."
fyne package --target android --app-id "$BUNDLE_ID"

echo "==> Installing on Android device..."
adb shell settings put global verifier_verify_adb_installs 0
adb install -r Sudoku.apk

echo "==> Cleaning up APK..."
rm -f Sudoku.apk

echo ""
echo "✅ Done! Find 'Sudoku' on your Android device."
