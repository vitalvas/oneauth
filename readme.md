# OneAuth

OneAuth is a simple authentication and authorization toolkit for works with Yubikey.

> **Note:**
> This project is still in development and not ready for production use.

## Roadmap

* Client
  * [x] SSH Agent
  * OS Support
    * [x] MacOS
    * [ ] Debian based distributions
  * Keys
    * [x] Insecure RSA 2048 (static key)
    * [x] Insecure ECC P-256 (static key)
    * [ ] Secure RSA 2048 (certificate based key with CA)
    * [ ] Secure ECC P-256 (certificate based key with CA)
* Server
  * [ ] CA Server
  * [ ] OTP Validation Server

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
