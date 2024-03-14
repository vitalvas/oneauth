# Installation

## Requirements

### MacOS

Required no less than 10.14 (Mojave).

No additional requirements. All required libraries are included in the OS.

#### Shell

For ZSH

```shell
echo "export SSH_AUTH_SOCK=${HOME}/.oneauth/ssh-agent.sock" >> ${HOME}/.zshrc
```

### Linux

#### Ubuntu / Debian

```bash
apt install libccid pcscd
systemctl enable pcscd.socket
systemctl restart pcscd.socket
```
