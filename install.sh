#!/bin/bash
# Author: CryoByte33 (original) / CryoUtils NG contributors (rewrite)

# Create a hidden directory for the script, if not present
mkdir -p "$HOME/.cryoutils_ng" &>/dev/null
cd "$HOME/.cryoutils_ng" || exit 1

# Download checksum to compare with local binary, if present
wget https://github.com/Lawlietr/cryoutils-ng/releases/download/latest/cu.md5 -O "$HOME/.cryoutils_ng/cu.md5" 2>&1
sleep 1
if md5sum -c --quiet cu.md5; then
  zenity --info --text="No update necessary!" --width=300
  exit 0
fi

# Uninstall swap resizer if present
# Delete legacy install directory
rm -rf "$HOME/.swap_resizer" &>/dev/null

# Remove legacy Desktop icons
rm -rf ~/Desktop/SwapResizerUninstall.desktop &>/dev/null
rm -rf ~/Desktop/SwapResizer.desktop &>/dev/null

# Remove old binary
rm -f "$HOME/.cryoutils_ng/cryoutils-ng" &>/dev/null

# Attempt to download the binary 3 times.
for i in {1..3}; do
  # Download binary
  wget https://github.com/Lawlietr/cryoutils-ng/releases/download/latest/cryoutils-ng -O "$HOME/.cryoutils_ng/cryoutils-ng" 2>&1 | sed -u 's/.* \([0-9]\+%\)\ \+\([0-9.]\+.\) \(.*\)/\1\n# Downloading at \2\/s, ETA \3/' | zenity --progress --title="Downloading CU2 Binary, attempt $i of 3..." --auto-close --width=500

  # Start a loop testing if zenity is running, and if not kill wget (allows for cancel to work)
  RUNNING=0
  while [ $RUNNING -eq 0 ]; do
    if [ -z "$(pidof zenity)" ]; then
      pkill wget
      RUNNING=1
    fi
    sleep 0.1
  done

  sleep 1
  # Compare checksum to new binary
  if md5sum -c --quiet cu.md5; then
    break 2
  fi

  if [ "$i" -ge "3" ]; then
    zenity --error --text="Install/upgrade of CryoUtils NG has failed!\n\nBinary couldn't be downloaded correctly, this may be a network or GitHub issue." --width=500
    exit 1
  fi
done

chmod +x "$HOME/.cryoutils_ng/cryoutils-ng"
rm -f cu.md5 &>/dev/null

# Remove old launcher
rm -f "$HOME/.cryoutils_ng/launcher.sh" &>/dev/null

# Install launcher script
wget https://raw.githubusercontent.com/Lawlietr/cryoutils-ng/main/launcher.sh -O "$HOME/.cryoutils_ng/launcher.sh"
chmod +x "$HOME/.cryoutils_ng/launcher.sh"

# Remove old icon
rm -f "$HOME/.cryoutils_ng/cryoutils-ng.png" &>/dev/null

# Install Icon
wget https://raw.githubusercontent.com/Lawlietr/cryoutils-ng/main/cmd/cryoutilities/Icon.png -O "$HOME/.cryoutils_ng/cryoutils-ng.png"
xdg-icon-resource install cryoutils-ng.png --size 64

# Create Desktop icons
rm -rf "$HOME"/Desktop/CryoUtilsNGUninstall.desktop 2>/dev/null
echo '#!/usr/bin/env xdg-open
[Desktop Entry]
Name=Uninstall CryoUtils NG
Exec=curl https://raw.githubusercontent.com/Lawlietr/cryoutils-ng/main/uninstall.sh | bash -s --
Icon=delete
Terminal=false
Type=Application
StartupNotify=false' >"$HOME"/Desktop/CryoUtilsNGUninstall.desktop
chmod +x "$HOME"/Desktop/CryoUtilsNGUninstall.desktop

rm -rf "$HOME"/Desktop/CryoUtilsNG.desktop 2>/dev/null
echo "#!/usr/bin/env xdg-open
[Desktop Entry]
Name=CryoUtils NG
Exec=bash $HOME/.cryoutils_ng/launcher.sh
Icon=cryoutils-ng
Terminal=false
Type=Application
StartupNotify=false" >"$HOME"/Desktop/CryoUtilsNG.desktop
chmod +x "$HOME"/Desktop/CryoUtilsNG.desktop

rm -rf "$HOME"/Desktop/UpdateCryoUtilsNG.desktop 2>/dev/null
echo "#!/usr/bin/env xdg-open
[Desktop Entry]
Name=Update CryoUtils NG
Exec=curl https://raw.githubusercontent.com/Lawlietr/cryoutils-ng/main/install.sh | bash -s --
Icon=bittorrent-sync
Terminal=false
Type=Application
StartupNotify=false" >"$HOME"/Desktop/UpdateCryoUtilsNG.desktop
chmod +x "$HOME"/Desktop/UpdateCryoUtilsNG.desktop

# Create Start Menu Icons
rm -rf "$HOME"/.local/share/applications/CryoUtilsNGUninstall.desktop 2>/dev/null
echo '#!/usr/bin/env xdg-open
[Desktop Entry]
Name=CryoUtils NG - Uninstall
Exec=curl https://raw.githubusercontent.com/Lawlietr/cryoutils-ng/main/uninstall.sh | bash -s --
Icon=delete
Terminal=false
Type=Application
Categories=Utility
StartupNotify=false' >"$HOME"/.local/share/applications/CryoUtilsNGUninstall.desktop
chmod +x "$HOME"/.local/share/applications/CryoUtilsNGUninstall.desktop

rm -rf "$HOME"/.local/share/applications/CryoUtilsNG.desktop 2>/dev/null
echo '#!/usr/bin/env xdg-open
[Desktop Entry]
Name=CryoUtils NG
Exec=bash $HOME/.cryoutils_ng/launcher.sh
Icon=cryoutils-ng
Terminal=false
Type=Application
Categories=Utility
StartupNotify=false' >"$HOME"/.local/share/applications/CryoUtilsNG.desktop
chmod +x "$HOME"/.local/share/applications/CryoUtilsNG.desktop

rm -rf "$HOME"/.local/share/applications/UpdateCryoUtilsNG.desktop 2>/dev/null
echo '#!/usr/bin/env xdg-open
[Desktop Entry]
Name=CryoUtils NG - Update
Exec=curl https://raw.githubusercontent.com/Lawlietr/cryoutils-ng/main/install.sh | bash -s --
Icon=bittorrent-sync
Terminal=false
Type=Application
Categories=Utility
StartupNotify=false' >"$HOME"/.local/share/applications/UpdateCryoUtilsNG.desktop
chmod +x "$HOME"/.local/share/applications/UpdateCryoUtilsNG.desktop

update-desktop-database ~/.local/share/applications

zenity --info --text="Install/upgrade of CryoUtils NG has been completed!" --width=300
