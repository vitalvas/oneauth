# Installation

## Requirements

### MacOS

Required no less than 10.14 (Mojave).

No additional requirements. All required libraries are included in the OS.

### Linux

#### Ubuntu / Debian

```bash
apt install libccid pcscd
systemctl enable pcscd.socket
systemctl restart pcscd.socket
```
