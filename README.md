# beacon

## Install
- Use goreleaser to make a dep package.
- `sudo apt install bluez bluez-firmware pi-bluetooth` Note a salt update will uninstall these packages.
- in `/lib/systemd/system/bluetooth.service` change the line `ExecStart=/usr/local/libexec/bluetooth/bluetoothd` to `ExecStart=/usr/local/libexec/bluetooth/bluetoothd --experimental`
- Comment out `dtoverlay=pi3-disable-bt` in `/boot/config.txt` RPI might need to be restarted after changing this. Note that a salt update will revert this change.

## Format of beacon

Len (bytes) | 1      | 1              | 2                       | 1       | 2        | 1    | variable | 4
------------|--------|----------------|-------------------------|---------|----------|------|----------|----
Data:       | length | AD Type (0xFF) | Manufacture ID (0x1212) | Version | DeviceID | type | data     | CRC


Data types     | hex value | Len      | Data
---------------|-----------|----------|--------------
Ping           | 0x01      | 0        | NA
Recording      | 0x02      | 0        | NA
Classification | 0x03      | Variable | number of classifications, (animal type, confidence) x number of classifications
Power off      | 0x04      | 2        | Minutes to sleep
