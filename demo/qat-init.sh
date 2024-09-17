#!/usr/bin/env bash
# This script is based on qatlib's qat_init.sh
NODE_NAME="${NODE_NAME:-}"
ENABLED_QAT_PF_PCIIDS=${ENABLED_QAT_PF_PCIIDS:-37c8 4940 4942 4944 4946}
DEVS=$(for pf in $ENABLED_QAT_PF_PCIIDS; do lspci -n | grep -e "$pf" | grep -o -e "^\\S*"; done)
SERVICES_LIST="sym asym sym;asym dc sym;dc asym;dc"
QAT_4XXX_DEVICE_PCI_ID="0x4940"
QAT_401XX_DEVICE_PCI_ID="0x4942"
QAT_402XX_DEVICE_PCI_ID="0x4944"
QAT_420XX_DEVICE_PCI_ID="0x4946"
SERVICES_ENABLED="NONE"
SERVICES_ENABLED_FOUND="FALSE"

AUTORESET_ENABLED="NONE"
AUTORESET_ENABLED_FOUND="FALSE"
AUTORESET_OPTIONS_LIST="on off"

check_config() {
  enabled_config="NONE"
  [ -f "conf/qat.conf" ] && enabled_config=$(grep "^$1=" conf/qat.conf | cut -d= -f 2 | grep '\S')
  [ -f "conf/qat-$NODE_NAME.conf" ] && enabled_config=$(grep "^$1=" conf/qat-"$NODE_NAME".conf | cut -d= -f 2 | grep '\S')
  is_config_valid="FALSE"
  if [ "$enabled_config" != "NONE" ]; then
    for config in $2
    do
      if [ "$config" = "$enabled_config" ]; then
        is_config_valid="TRUE"
        break
      fi
    done
  fi
  if [ "$is_config_valid" = "TRUE" ]; then
    echo "$enabled_config"
  else
    echo "NONE"
  fi
}

is_found() {
  if [ "$1" = "NONE" ]; then
    echo "FALSE"
  else
    echo "TRUE"
  fi
}

sysfs_config() {
  if [ "$SERVICES_ENABLED_FOUND" = "TRUE" ]; then
    for dev in $DEVS; do
      DEVPATH="/sys/bus/pci/devices/0000:$dev"
      PCI_DEV=$(cat "$DEVPATH"/device 2> /dev/null)
      if [ "$PCI_DEV" != "$QAT_4XXX_DEVICE_PCI_ID" ] && [ "$PCI_DEV" != "$QAT_401XX_DEVICE_PCI_ID" ] && [ "$PCI_DEV" != "$QAT_402XX_DEVICE_PCI_ID" ] && [ "$PCI_DEV" != "$QAT_420XX_DEVICE_PCI_ID" ]; then
        continue
      fi

      CURRENT_SERVICES=$(cat "$DEVPATH"/qat/cfg_services)
      if [ "$CURRENT_SERVICES" != "$SERVICES_ENABLED" ]; then
        CURRENT_STATE=$(cat "$DEVPATH"/qat/state)
        if [ "$CURRENT_STATE" = "up" ]; then
          echo down > "$DEVPATH"/qat/state
        fi
        echo "$SERVICES_ENABLED" > "$DEVPATH"/qat/cfg_services
        CURRENT_SERVICES=$(cat "$DEVPATH"/qat/cfg_services)
      fi
      echo "Device $dev configured with services: $CURRENT_SERVICES"
    done
  fi
}

enable_sriov() {
  for dev in $DEVS; do
  DEVPATH="/sys/bus/pci/devices/0000:$dev"
  NUMVFS="$DEVPATH/sriov_numvfs"
  if ! test -w "$NUMVFS"; then
    echo "error: $NUMVFS is not found or not writable. Check if QAT driver module is loaded"
    exit 1
  fi
  if [ "$(cat "$NUMVFS")" -ne 0 ]; then
    echo "$DEVPATH already configured"
  else
    tee "$NUMVFS" < "$DEVPATH/sriov_totalvfs"
  fi
  done
}

enable_auto_reset() {
  if [ "$AUTORESET_ENABLED_FOUND" = "TRUE" ]; then
    for dev in $DEVS; do
      DEVPATH="/sys/bus/pci/devices/0000:$dev"
      AUTORESET_PATH="$DEVPATH/qat/auto_reset"
      if ! test -w "$AUTORESET_PATH"; then
        echo "error: $AUTORESET_PATH is not found or not writable. Check if QAT driver module is loaded. Skipping..."
        exit 1
      fi
      if [ "$(cat "$AUTORESET_PATH")" == "$AUTORESET_ENABLED" ]; then
        echo "$DEVPATH's auto reset is already $AUTORESET_ENABLED"
      else
        echo "$AUTORESET_ENABLED" > "$AUTORESET_PATH" && echo "$DEVPATH's auto reset has been set $AUTORESET_ENABLED"
      fi
    done
  fi
}

SERVICES_ENABLED=$(check_config "ServicesEnabled" "$SERVICES_LIST")
SERVICES_ENABLED_FOUND=$(is_found "$SERVICES_ENABLED")
sysfs_config
enable_sriov

AUTORESET_ENABLED=$(check_config "AutoresetEnabled" "$AUTORESET_OPTIONS_LIST")
AUTORESET_ENABLED_FOUND=$(is_found "$AUTORESET_ENABLED")
enable_auto_reset
