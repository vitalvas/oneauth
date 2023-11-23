# Roadmap

## Client

* [ ] Self update
* [ ] Config file
* [ ] Encryption and decryption of data
* [ ] Signing and verification of data
* [ ] Using PIN policy
* [x] Using touch policy

### Agent

* [x] SSH Agent
* [ ] Check for correct source requester to unix socket (deny access from another user)
* [ ] RPC Server for cuncurrent access to Yubikey

### OS Support

* [x] MacOS
  * [x] Arch: amd64
  * [ ] Arch: arm64
  * [ ] Using `launchd` for agent
  * [ ] Using keychain for storing PIN and PUK
* [ ] Debian based distributions

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
* [ ] Enable/disable interfaces for USB (OTP, PIV, FIDO2, FIDO U2F, OATH, OpenPGP, ...)

### Keys

* [x] Insecure RSA 2048 (static key)
* [x] Insecure ECC P-256 (static key)
* [ ] Secure RSA 2048 (certificate based key with CA)
* [ ] Secure ECC P-256 (certificate based key with CA)
* [ ] PIV Certificates
  * [ ] Authentication
  * [ ] Digital Signature
  * [ ] Key Management
  * [ ] Card Authentication

## RoboClient

SSH Agent without using Yubikey. Based on short lived certificates (for example: 1 day).

Usage: give access to server for robots (CI/CD, backup, ...)

* [ ] Config file
* [ ] SSH Agent

## Server

* [ ] CA Server
* [ ] OTP Validation Server

## Test SSH Server

Usage: test SSH Agent and Yubikey
