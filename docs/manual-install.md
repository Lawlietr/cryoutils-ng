# Manual Install — CryoUtils NG

## Prerequisites

1. Download both binaries from the [releases page](https://github.com/Lawlietr/cryoutils-ng/releases/tag/latest):
   - `cryoutils-ng` (CLI)
   - `cryoutils-ng-desktop` (desktop web server / GUI)
2. Download `icon.png` from the repository root

## Steps

1. Create the install directory:
   ```bash
   mkdir ~/.cryoutils_ng
   ```

2. Move both binaries and the icon into the install directory:
   ```bash
   cd ~/.cryoutils_ng
   chmod +x cryoutils-ng cryoutils-ng-desktop
   xdg-icon-resource install icon.png --size 64
   ```

3. Download `launcher.sh` from the repository:
   ```bash
   curl -L https://raw.githubusercontent.com/Lawlietr/cryoutils-ng/main/launcher.sh -o launcher.sh
   chmod +x launcher.sh
   ```

4. Create Desktop icons:
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

5. Create Start Menu icons:
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

6. Update desktop database:
   ```bash
   update-desktop-database ~/.local/share/applications
   ```

## Desktop Server (GUI)

Double-clicking the "CryoUtils NG" icon launches `cryoutils-ng-desktop`, which starts a localhost web server and opens your browser automatically. The UI is served at `http://127.0.0.1:<PORT>/?token=<TOKEN>`.

If you prefer to run the desktop server manually:

```bash
~/.cryoutils_ng/cryoutils-ng-desktop
```

## CLI Usage

```bash
sudo ~/.cryoutils_ng/cryoutils-ng status
sudo ~/.cryoutils_ng/cryoutils-ng recommended
sudo ~/.cryoutils_ng/cryoutils-ng stock
sudo ~/.cryoutils_ng/cryoutils-ng help
```

Now, you should be all set to use CryoUtils NG as normal!
