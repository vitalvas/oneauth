# Roadmap

## Client

* [ ] Self update
    * [x] Check for new version
    * [ ] Download and install new version
    * [ ] Periodic check for new version and notify user
* [x] Config file
* [ ] Using PIN policy
    * [ ] Ask user for PIN
    * [x] Store PIN in OS keyring
* [x] Using touch policy
* [ ] Using setup profile from server
* [x] Initial setup
    * [x] Full reset and setup
    * [ ] Partial setup (only secure keys)

### Agent

* [x] SSH Agent
* [x] Check for correct source requester to unix socket (deny access from another user)
* [ ] RPC Server for cuncurrent access to Yubikey
* [ ] Write audit log

### OS Support

* [x] MacOS
    * [x] Arch: amd64
    * [x] Arch: arm64
    * [x] Using `launchd` for agent
    * [x] Using keychain for storing PIN and PUK
* [ ] Linux
    * [x] Arch: amd64
    * [ ] Arch: arm64
    * [ ] Using `systemd` for agent
    * [ ] Debian based distributions
* [ ] Windows (not sure if this support is needed at all... I don't have a place to test it)

### Yubikey

* [x] Reset Yubikey to factory defaults
    * [x] Reset PIV applet
    * [ ] Reset OTP applet
    * [ ] Reset FIDO2 applet
* [x] Change PIN
* [x] Change PUK
* [x] Unlock PIN using PUK
* [ ] Rotate insecure keys
* [ ] Rotate secure keys
* [ ] Enable/disable interfaces for USB/NFC (OTP, PIV, FIDO2, FIDO U2F, OATH, OpenPGP, ...)

### Keys (PIV applet)

* [x] Insecure RSA 2048 (static key)
* [x] Insecure ECC P-256/P-384 (static key)
* [ ] Secure RSA 2048 (certificate based key with CA)
* [ ] Secure ECC P-256/P-384 (certificate based key with CA)
* [ ] PIV Certificates
    * [ ] Authentication
    * [ ] Digital Signature
    * [ ] Key Management
    * [ ] Card Authentication

## RoboClient

SSH Agent without using Yubikey. Based on short lived certificates (for example: 1 day).

Usage: give access to server for robots (CI/CD, backup, automation,...)

* [ ] Config file
* [ ] SSH Agent

## Server

* [ ] CA Server (PKI)
* [ ] Serve setup profiles for clients
* [x] OTP Validation Server
    * [ ] Validate user key ownership
    * [x] Validate OTP
    * [x] Validate OTP with YubiCloud
    * [ ] Audit log
* [ ] User directory
    * [ ] User management
    * [ ] Sync from external directory (LDAP, Active Directory, scripts, ...)
* [ ] Support [YubiHSM 2](https://www.yubico.com/us/product/yubihsm-2/) (please, donate one or two to me)

## Test SSH Server

Usage: test SSH Agent and Yubikey

## Radius Server

Usage: give access to network for users
