# Client

* [ ] Self update
* [x] Config file
* [ ] Using PIN policy
* [x] Using touch policy

## Agent

* [x] SSH Agent
* [x] Check for correct source requester to unix socket (deny access from another user)
* [ ] RPC Server for cuncurrent access to Yubikey

## OS Support

* [x] MacOS
    * [x] Arch: amd64
    * [x] Arch: arm64
    * [ ] Using `launchd` for agent
    * [x] Using keychain for storing PIN and PUK
* [ ] Linux
    * [ ] Arch: amd64
    * [ ] Arch: arm64
    * [ ] Debian based distributions
* [ ] Windows

## Yubikey

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

## Keys (PIV applet)

* [x] Insecure RSA 2048 (static key)
* [x] Insecure ECC P-256 (static key)
* [ ] Secure RSA 2048 (certificate based key with CA)
* [ ] Secure ECC P-256 (certificate based key with CA)
* [ ] PIV Certificates
    * [ ] Authentication
    * [ ] Digital Signature
    * [ ] Key Management
    * [ ] Card Authentication
