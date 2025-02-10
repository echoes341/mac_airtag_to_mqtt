#!/bin/bash
set -euo pipefail

sed -e "s|\$REPOPATH|$PWD|g" mac_airtag_to_mqtt.plist \
  | sudo tee /Library/LaunchDaemons/com.echoes341.mac_airtag_to_mqtt.plist
sudo launchctl load -w /Library/LaunchDaemons/com.echoes341.mac_airtag_to_mqtt.plist
