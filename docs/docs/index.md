---
hide:
    - toc
---
# OneAuth

OneAuth is a custom SSH agent (maybe not only in the future) for working with built-in PKI on top of Yubikey, which:

* does initial Yubikey setup
* generates and stores SSH keys and certificates on Yubikey
* provides access to them via ssh agent protocol over:
    * UNIX-socket (used and supported by the vast majority of SSH clients under GNU/Linux and macOS)

## Supported Devices

These are the devices that are currently tested and supported:

* YubiKey 5 Nano *(recommended)*
* YubiKey 5C Nano *(recommended)*
* YubiKey 5 NFC

Due to limitations in supporting functionality in the keys themselves, only version 5 is supported.
