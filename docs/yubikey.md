# YubiKey

## Setup

### Touch Policy

* `always` - always require touch
* `cached` - require touch and is cache for 15 seconds
* `never` - never require touch

## YubiKey tools

Status:

```bash
yubico-piv-tool -a status
```

Read certificate:

```bash
yubico-piv-tool -a read-certificate -s 95
```
