# CryoUtils NG

A Steam Deck utility to manage swap files, swappiness, and memory parameters — a rewrite of the original [CryoUtilities](https://github.com/CryoByte33/steam-deck-utilities) by CryoByte33.

> **Note**: This is a derivative work of CryoUtilities (unmaintained since 2023). It is distributed under the same GPLv3 license. This project coexists with the original CryoUtilities — it does not overwrite the original install.

## Features

* One-click set-to-recommended settings
* One-click revert-to-stock settings
* Swap Tuner
    * Swap File Resizer + Recovery
    * Swappiness Changer
* Memory Parameter Tuning
    * HugePages Toggle
    * Compaction Proactiveness Changer
    * HugePage Defragmentation Toggle
    * Page Lock Unfairness Changer
    * Shared Memory (shmem) Toggle
* Storage Manager
    * Sync shadercache and compatdata to the same location the game is installed
    * Delete shadercache and compatdata for whichever games you select
    * Delete the shadercache and compatdata for all uninstalled games with a single click
* Full CLI mode
* Web UI (desktop mode)

## Install

### Simple (Recommended)

Download the [InstallCryoUtilsNG.desktop](https://raw.githubusercontent.com/Lawlietr/cryoutils-ng/main/InstallCryoUtilities.desktop) file to your desktop (right click and save file) on your Steam Deck, remove the `.download` from the end of the file name, then double-click it.

This will install both the CLI and desktop server binaries, create desktop icons, and create menu entries.

### Manual

See [manual-install.md](docs/manual-install.md).

## Usage

**NOTE**: This **REQUIRES** a password set on the Steam Deck. That can be done with the `passwd` command.

### GUI (Desktop Mode)

After installation, double-click the "CryoUtils NG" icon on the desktop or find it in the application menu under "Utilities". This launches the desktop web server and opens your browser automatically as a chromeless app window (`--new-window --app=<url>`).

The web UI runs on `127.0.0.1` with a per-launch random token. No network exposure — it only listens on localhost.

#### Browser Launch Fallback Chain

The desktop server tries browsers in this order:
1. Native Chromium-family browsers (`google-chrome`, `chromium`, `chromium-browser`, `microsoft-edge`, `brave-browser`, `brave`, `firefox`) via `--new-window --app=`
2. Flatpak browsers (`com.brave.Browser`, `org.chromium.Chromium`, `com.google.Chrome`, `com.microsoft.Edge`)
3. `steam://openurl` (Steam's built-in CEF browser — guaranteed to exist on Deck)
4. `xdg-open` (last resort)

#### CLI Flags

```bash
# Print the URL without opening a browser
cryoutils-ng-desktop -no-browser

# Force a specific browser path
cryoutils-ng-desktop -browser /usr/bin/chromium
```

### CLI

```
sudo ~/.cryoutils_ng/cryoutils-ng <command> [parameter]
```

If you want to see the available commands and accepted values, you can use:

```
sudo ~/.cryoutils_ng/cryoutils-ng help
```

**Note**: You _need_ to use sudo for the tweaks to work, otherwise it can't write to the necessary locations on disk.

### Status Command

Check current settings via CLI:

```
sudo ~/.cryoutils_ng/cryoutils-ng status
```

### Building the Desktop Server from Source

The desktop server binary is not distributed via GitHub Releases. To build it:

```bash
# Build web UI first
cd web && npm ci && npm run build

# Build desktop server (static binary, no CGO)
cd ..
CGO_ENABLED=0 go build -o cryoutils-ng-desktop ./cmd/desktop
```

Then run it directly:

```bash
./cryoutils-ng-desktop
```

It will print a URL like `http://127.0.0.1:PORT/?token=TOKEN` — open that in your browser.

Or use flags:

```bash
# Print URL only (for scripting or manual copy)
./cryoutils-ng-desktop -no-browser

# Force a specific browser
./cryoutils-ng-desktop -browser /usr/bin/brave-browser
```

## Upgrade

Double-click the "Update CryoUtils NG" icon on the desktop, you will get a dialog box when the update is complete.

## Uninstall

Double-click the "Uninstall CryoUtils NG" icon on the desktop, you will be asked if you're sure, then asked if you want to revert the tweaks that have been made.

## Revert To Default Settings

To revert to the Steam Deck defaults, do one of the following:

* Launch CryoUtils NG (desktop mode) and click "Stock" on the homepage.
* Uninstall CryoUtils NG, you'll be asked if you want to revert to stock settings. Choose yes.

After choosing these options, the Deck will be identical to an unmodified version.

## Known Issues

* If the drive becomes full during the swap file resize, you can trigger a known SteamOS bug that causes boot loops.
    * CryoUtils NG is programmed in such a way to not allow this, but in the very worst cases it's still possible if
      something is operating/downloading in the background, at the same time CryoUtils NG resizes the swap file.
    * In the event that it happens, you need to either get into a live environment and delete some files, or reinstall
      SteamOS with the non-destructive method.
* While using CLI mode, it is possible that the swap file resize takes long enough that the sudo credentials will time
  out.
    * This does not occur in GUI mode, due to how authentication is implemented.

## FAQ

See [the FAQ page](docs/faq.md).

## What does it do?

See [the tweak explanation page](docs/tweak-explanation.md).

## Troubleshooting

### CryoUtils NG doesn't appear after double-clicking the icon on the desktop

* Make sure that you're using SteamOS 3.4 or later
* Verify that `/home/deck/.cryoutils_ng/cryoutils-ng` and `/home/deck/.cryoutils_ng/cryoutils-ng-desktop` exist
    * `/home/deck/.cryoutils_ng` is a hidden directory, so ensure that you can view hidden files

### The desktop server won't start

* Make sure both `cryoutils-ng` and `cryoutils-ng-desktop` exist in `~/.cryoutils_ng/`
* Check the log at `~/.cryoutils_ng/cryoutils_ng.log`

## Attribution

This project is a rewrite of [CryoUtilities](https://github.com/CryoByte33/steam-deck-utilities) by CryoByte33, which has been unmaintained since 2023. The core tuning logic is preserved and rewritten in Go; the UI is independently designed as a single-page web application.

## License

This project is distributed under the GNU General Public License v3.0 (GPLv3), inherited from the original CryoUtilities project. See [LICENSE](LICENSE) for details.
