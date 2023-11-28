---
hide:
    - toc
---
# OneAuth

!!! danger
    This project is in the early stages of development and is not ready for use in production.

OneAuth is a custom SSH agent (maybe not only in the future) for working with built-in PKI on top of Yubikey, which:

* does initial Yubikey setup
* generates and stores SSH keys and certificates on Yubikey
* provides access to them via ssh agent protocol over:
    * UNIX-socket (used and supported by the vast majority of SSH clients under GNU/Linux and macOS)

## Supported Devices

!!! warning
    Due to limitations in supporting functionality in the keys themselves, **only 5 Series** is supported.

These are the devices that are currently tested and supported:

* [YubiKey 5 Nano](https://www.yubico.com/product/yubikey-5-nano/) *(recommended)*
* [YubiKey 5C Nano](https://www.yubico.com/product/yubikey-5c-nano/) *(recommended)*
* [YubiKey 5 NFC](https://www.yubico.com/product/yubikey-5-nfc/)
