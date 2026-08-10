#!/bin/bash
# Author: CryoByte33 (original) / CryoUtils NG contributors (rewrite)

# This is nested for good reason, zenity won't exit the entire script if the 'x' button is pressed.
# Nesting it forces execution only if an option is selected.
if zenity --question --title="Disclaimer" --text="This script will uninstall CryoUtils NG.\n\n<b>Disclaimer:</b> Do you want to proceed?" --width=600 2>/dev/null; then
  if zenity --question --title="Revert" --text="Do you want to revert the tweaks made by CryoUtils NG?\n\n<b>Note:</b> This does NOT move the game data to the original location on the SSD." --width=600 2>/dev/null; then
    # Ask for password
    hasPass=$(passwd -S "$USER" | awk -F " " '{print $2}')
    if [[ $hasPass != "P" ]]; then
      zenity --error --title="Password Error" --text="Password is not set, please set one in the terminal with the <b>passwd</b> command, then run this again." --width=400 2>/dev/null
      exit 1
    fi
    PASSWD="$(zenity --password --title="Enter Password" --text="Enter Deck User Password (not Steam account!)" 2>/dev/null)"
    echo "$PASSWD" | sudo -v -S
    ans=$?
    if [[ $ans == 1 ]]; then
      zenity --error --title="Password Error" --text="Incorrect password provided, please run this command again and provide the correct password." --width=400 2>/dev/null
      exit 1
    fi
    # Revert everything to stock
    sudo bash "$HOME"/.cryoutils_ng/cryoutils-ng stock
  fi
  # Delete install directory
  rm -rf "$HOME/.cryoutils_ng"

  # Remove Desktop icons
  rm -rf "$HOME"/Desktop/CryoUtilsNGUninstall.desktop 2>/dev/null
  rm -rf "$HOME"/Desktop/CryoUtilsNG.desktop 2>/dev/null
  rm -rf "$HOME"/Desktop/UpdateCryoUtilsNG.desktop 2>/dev/null

  # Remove Start Menu shortcuts
  rm -rf "$HOME"/.local/share/applications/CryoUtilsNGUninstall.desktop 2>/dev/null
  rm -rf "$HOME"/.local/share/applications/CryoUtilsNG.desktop 2>/dev/null
  rm -rf "$HOME"/.local/share/applications/UpdateCryoUtilsNG.desktop 2>/dev/null
  update-desktop-database ~/.local/share/applications

  # Remove icon from KDE
  xdg-icon-resource uninstall cryoutils-ng 2>/dev/null
fi
