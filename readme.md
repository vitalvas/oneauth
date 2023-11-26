# OneAuth

> **Note:**
> This project is still in development and not ready for production use.

OneAuth is a simple authentication toolkit for works with Yubikey.

* [Roadmap](https://oneauth.vitalvas.dev/roadmap/)

## Supported Devices

These are the devices that are currently tested and supported:

* YubiKey 5 Nano *(recommended)*
* YubiKey 5C Nano *(recommended)*
* YubiKey 5 NFC

Due to limitations in supporting functionality in the keys themselves, only version 5 is supported.

## Generated Keys

* RSA 2048 - named `insecure-rsa`. (slot 0x95)
* ECC P-256 - named `insecure-ecdsa`. (slot 0x94)
