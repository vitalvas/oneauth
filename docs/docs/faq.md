# FAQ

## Can I store OTP codes?

Maybe. You can store TOTP codes in YubiKey. TOTP/HOTP codes are stored inside the OATH applet. In the future, OATH reset will be added.

## Can I use for WebAuthn?

Maybe. In the future, FIDO2 and U2F reset will be added. After the reset, you will need to add the key as a new one.

## Where are certificates and keys stored?

The certificates and keys are stored in the PIV applet. The PIV applet is a smart card applet that can store certificates and keys. The PIV applet is a standard applet that is supported by many OS.

## Can I store my certificates?

Not recommended. The management process can delete them at any time.

## Can I export my private key?

No. The private key is stored inside the YubiKey and cannot be exported.
