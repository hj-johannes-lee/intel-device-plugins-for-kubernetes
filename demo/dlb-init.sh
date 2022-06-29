#!/bin/bash -eu

enable_and_configure_vfs() {
  devpath=$1
  num_vfs_per_pf=$2
  # enable sriov
  sriov_numvfs_path="$devpath/sriov_numvfs"
  if ! test -w "$sriov_numvfs_path"; then
    echo "error: $sriov_numvfs_path is not found or not writable. Check if dlb driver module is loaded."
  else
    if [ "$(cat "$sriov_numvfs_path")" -ne 0 ]; then
      echo "$devpath already configured"
    else
      echo -n "$num_vfs_per_pf" > "$sriov_numvfs_path"
      # configure vfs
      for virtfnN in $(basename -a "$devpath"/virtfn*) ; do
        #unbind vf
        vf_pciid=$(basename "$(realpath "$devpath/$virtfnN")")
        dlb_pci_driver_path=/sys/bus/pci/drivers/dlb2
        echo -n "$vf_pciid" > $dlb_pci_driver_path/unbind

        #configure vf
        vf_resources_path="$devpath/vf${virtfnN#virtfn}_resources"
        targets="num_atomic_inflights num_dir_credits num_dir_ports num_hist_list_entries num_ldb_credits num_ldb_ports num_ldb_queues num_sched_domains num_sn0_slots num_sn1_slots"
        for target in $targets; do
          val=$(cat "$devpath/total_resources/$target")
          share=$((val/NUM_VFS_PER_PF))
          echo "$share" | tee -a "$vf_resources_path/$target"
        done

        #bind vf back to dlb2 driver
        echo -n "$vf_pciid" > $dlb_pci_driver_path/bind
      done
      echo "$devpath configured"
    fi
  fi
}

#get pci ids of pfs
DEVS=$(realpath /sys/bus/pci/drivers/dlb2/????:??:??\.0)
NUM_VFS_PER_PF=${NUM_VFS_PER_PF:-1}

skip=true
for dev in $DEVS; do
  #every 2nd pf is used for configuring vfs
  if $skip ; then
    skip=false
    echo "$dev" "skipped"
  else
    skip=true
    enable_and_configure_vfs "$dev" "$NUM_VFS_PER_PF"
  fi
done
