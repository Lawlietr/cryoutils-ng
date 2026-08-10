# Manual Install — CryoUtils NG

## Prerequisites

1. Install Go 1.26+ on the Steam Deck (or cross-compile from another machine)
2. Download the binary from the releases page

## Steps

1. Create the install directory:
   ```bash
   mkdir ~/.cryoutils_ng
   ```

2. Go to [the releases page](https://github.com/Lawlietr/cryoutils-ng/releases/tag/latest) and download `cryoutils-ng`.

3. Go to [launcher.sh](https://github.com/Lawlietr/cryoutils-ng/blob/main/launcher.sh), right click on "Raw" and "Save Link As" to the downloads folder.

4. Go to [icon.png](https://github.com/Lawlietr/cryoutils-ng/blob/main/icon.png), right click on "Raw" and "Save Link As" to the downloads folder naming it `cryoutils-ng.png`.

5. Move all 3 downloaded files to `/home/deck/.cryoutils_ng`:
   ```bash
   cd ~/.cryoutils_ng
   chmod +x cryoutils-ng
   xdg-icon-resource install cryoutils-ng.png --size 64
   ```

6. Create Desktop icons:
   ```bash
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
   ```

7. Create Start Menu icons:
   ```bash
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
   ```

8. Update desktop database:
   ```bash
   update-desktop-database ~/.local/share/applications
   ```

Now, you should be all set to use CryoUtils NG as normal!
