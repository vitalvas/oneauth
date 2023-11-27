# OneAuth

OneAuth is a custom SSH agent (maybe not only in the future) for working with built-in PKI on top of Yubikey, which:

* does initial Yubikey setup
* generates and stores SSH keys and certificates on Yubikey
* provides access to them via ssh agent protocol over:
    * UNIX-socket (used and supported by the vast majority of SSH clients under GNU/Linux and macOS)
