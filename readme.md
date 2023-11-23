# OneAuth

OneAuth is a simple authentication toolkit for works with Yubikey.

* [Roadmap](docs/roadmap.md)

> **Note:**
> This project is still in development and not ready for production use.

## Supported Devices

These are the devices that are currently tested and supported:

* YubiKey 5 Nano *(recommended)*
* YubiKey 5C Nano *(recommended)*
* YubiKey 5 NFC

Due to limitations in supporting functionality in the keys themselves, only version 5 is supported.

## Generated Keys

* RSA 2048 - named `insecure-rsa`. (slot 0x95)
* ECC P-256 - named `insecure-ecdsa`. (slot 0x94)

## Installation

### Requirements

#### MacOS

Required no less than 10.14 (Mojave).

## Usage

List all available keys:

```bash
SSH_AUTH_SOCK=~/.oneauth/ssh-agent.sock ssh-add -L
```

SSH client configuration:

```bash
Host *
  # for authentication with Yubikey
  IdentityAgent ~/.oneauth/ssh-agent.sock
  # for forward keys to remote host (usage for bastion or jump host)
  ForwardAgent ~/.oneauth/ssh-agent.sock
```
