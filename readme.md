# OneAuth

OneAuth is a simple authentication and authorization toolkit for works with Yubikey.

## Supported Devices

These are the devices that are currently tested and supported:

* YubiKey 5 Nano
* YubiKey 5C Nano

## Generated Keys

* 1st slot: RSA 2048 - named `insecure-rsa`. Touch policy is set to `never`. (slot 0x95)
* 2nd slot: ECC P-256 - named `insecure-ecdsa`. Touch policy is set to `never`. (slot 0x94)

## Installation

### Requirements

#### MacOS

Required no less than 10.14 (Mojave).

#### Debian based distributions

```bash
apt install libpcsclite-dev
```

## Usage

```bash
SSH_AUTH_SOCK=~/.oneauth/ssh-agent.sock ssh-add -L
```

```bash
Host *
    IdentityAgent ~/.oneauth/ssh-agent.sock
```
