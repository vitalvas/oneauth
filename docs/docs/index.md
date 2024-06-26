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

!!! warning
    Due to limitations in supporting functionality in the keys themselves, **only 5 Series** is supported.

These are the devices that are currently tested and supported:

* [YubiKey 5 Nano](https://www.yubico.com/product/yubikey-5-nano/?utm_source=oneauth.vitalvas.dev) *(recommended)*
* [YubiKey 5C Nano](https://www.yubico.com/product/yubikey-5c-nano/?utm_source=oneauth.vitalvas.dev) *(recommended)*
* [YubiKey 5C](https://www.yubico.com/product/yubikey-5c/?utm_source=oneauth.vitalvas.dev)
* [YubiKey 5 NFC](https://www.yubico.com/product/yubikey-5-nfc/?utm_source=oneauth.vitalvas.dev)
