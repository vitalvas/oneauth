# How To

If you don't understand why OneAuth is focused on hardware tokens and certificates, this section should clear things up a bit.

## The private part of the key is not exportable

As long as the user has access to the private part of the key (whether encrypted or not), it is regularly transferred between work/personal devices, etc. Encryption does not help in any way with copy protection.

All this greatly increases the attack surface, since there are no guarantees that the user’s SSH key will not be found by someone (be it an “internal intruder” or malware).

In the case of a hardware token, the keys will be generated inside the token and it cannot be obtained from there by any means.
It cannot be extracted, stored on the file system, copied anywhere, stolen by malware, etc.

## Need be persistence for an attacker

Due to the inability to extract the private part of the key (see above) and use it from another location, the attacker will need to organize persistence on the user’s workstation.

I.e. To perform any actions on behalf of the user, the attacker must “sit” on his workstation, learn to survive reboots, etc.
On our part, this greatly increases the likelihood of detecting such an infection, but for the attacker, it greatly narrows the window of opportunity.

## Certificates or private keys

The use of certificates provides many different advantages in the context of security:

* Check the signature tree for validity. (i.e. check that the certificate is signed by a trusted CA)
* The ability to revoke a certificate
* The ability to control the validity period of the certificate
* The ability to control the use of the certificate (e.g. only for authentication, only for signing, etc.)
* No need to deploy each key certificate to the server (you only need trust in the CA)
* Privileges can be inserted into the certificate.

Typically, private keys are an obscure entity in a security context. The use of private keys should be clearly regulated, and used only where there is no support for the use of certificates.

## The passwordless

Using certificates allows you to authenticate on the server without any passwords.

Authentication using passwords brings many different problems.

For example, if we do this using Active Directory using sssd. This raises two problems:

* the user enters his password on a potentially hacked server. If we imagine a situation where an attacker hacked the server:
    * SRE/DevOps comes to her to sort out the problem
    * enters his domain password
    * the attacker receives the password to further develop the attack
* It allows you to brute domain passwords. Since this is a server segment, monitoring and investigating such incidents is very complicated

## Why yubikey?

We considered several options:

* Yubikey
* TPM in laptops and desktops
* Secure Enclave in Apple devices

Our requirements for a good hardware token:

* The ability to generate keys inside the token
* The key must not be able to export the private key (by any means)
* Support at least ECDSA and RSA keys. (Ed25519 is a plus)
* You can buy it (get it) anywhere (easy availability)
* The price should not be too high (up to $100)
* The ability to use it on any device (not tied to a specific device)
* The ability to use it on any OS (not tied to a specific OS)
* Easy integration
* Allows you to attest a stored key (i.e. check that the key is stored in the token and has not been exported from there)
* The ability to use it as a smart card
* Complies with NIST smart card security requirements

If you can teach TPM or Secure Enclave to fulfill our requirements, please tell us :)

## Why exactly this set of requirements?

### ECDSA keys

ECDSA is a modern algorithm that is not inferior to RSA in terms of security, but is much more compact in terms of key size.

They have good support from both OpenSSH and a variety of tokens.

Safety. For example, RSA requires increasingly longer key lengths and will likely be disabled in the future and, in the long term, support will be removed.

### Touch policy

We need a touch policy (literally, whether it is necessary to physically touch the token to unlock the key) to complicate life not only for the user (:smile:) , but also for the attacker.

The point is this: imagine that the attacker is on the same machine (be it a workstation or a server) with an SSH agent in which the token’s private keys are available:

* without touch policy: went to the server via SSH, signed a challenge in the agent, got to the server
* with touch policy: trick the user into touching the token or wait for the next time he goes to the server -> smaller window for attack + more noisy -> greatly increases the likelihood of detecting such an infection

And due to the fact that touch policy is usually configured for each key separately, we can control how secure the key storage must be in order to enter a particular environment.

For example, to access some endpoints that are described in NIST, there is no need to enable PIN and/or Touch policy. (or does it directly indicate what needs to be turned on or off)

When using Yubikey, the following policies are available:

* `always` - always require touch
* `cached` - require touch and caches it for 15 seconds (default)
* `never` - do not ask for touch

The policy is assigned to each slot (certificate) individually.

### Attestation

Attestation is a mechanism that allows the server side (PKI) to gain knowledge of how a key is stored and/or what its access policy is.
Typically it looks like receiving a certificate with key information that is written on a token certificate that is signed by the vendor's certificate.
Thus, using yubikey as an example, the server side can find out that:

* the key is actually stored in yubikey
* what is its pinning/touching policy

And depending on this knowledge, make decisions about trust and access to a particular environment. In our case - on which CA to issue a certificate to the user.

## Why is an external hardware token still secure?

Sometimes you may worry whether the stolen token can be used on another machine. We answer - in theory it is possible, in practice:

* For all keys we require the use of a PIN (including checking during certification) -> if you steal a yubikey you need to know the PIN
* There are three attempts to enter a PIN in yubikey, after that you can only reset the PIN using PUK
* When entering PUK there are three attempts, after that only a complete reset of PIV (loss of all private keys and their certificates)

Disclaimer: it is assumed that if you have disk encryption disabled, then you have much bigger problems than working with yubikey.

## Where is the world heading?

To a greater extent, everyone is moving in the direction of:

### Use SSH CA

They switch to an SSH CA, some more categorically (e.g. sew an ACL directly into the certificate and on the remote side unconditionally trust it), others less so, using the certificate only for authentication.

It is important to understand that usually, in order to manage access at the certificate level, you need to either have a custom SSH client or small infrastructure, or completely tie up the ability to access the product on an enroll service.

### Implement hardware tokens

Here you need to understand that in addition to certificates, many other different authentication methods have been invented.

Most of them look like simple generation of private/public keys without centralized management. (the output looks like a regular SSH key)

Mature companies integrate with CA.

### Build bastions (aka jump hosts)

Counting on what they can:

* control access to the bastion
* control access from the bastion to the target host (and to end hosts bypassing bastion)
* monitoring and session logging
* implement MFA (FIDO2? -> hardware token, PUSH? -> DUO, ...)

This is an additional measure to a correct authentication method, and not a replacement for it.
