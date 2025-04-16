#!/bin/bash

# Find all overlay mounts and unmount them
mount | grep overlay | awk '{print $3}' | while read mountpoint; do
  echo "Unmounting $mountpoint"
  umount "$mountpoint"
done