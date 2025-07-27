# YubiKey KSM Server

## What It Does

Validates YubiKey OTPs without using Yubico's cloud service. Stores
YubiKey AES keys and decrypts OTPs locally.

## Quick Start

1. **Configure** (`config.yaml`):

   ```yaml
   server:
     address: localhost:8002
   
   database:
     type: sqlite
     sqlite:
       path: /var/lib/oneauth/yubikey_ksm.db
   
   security:
     master_key: "your-32-character-master-key-here"
   
   logging:
     level: info
     format: json
   ```

2. **Run**:

   ```bash
   ./oneauth-yubikey-ksm-server --config config.yaml
   ```

3. **Test**:

   ```bash
   curl http://localhost:8002/api/v1/health
   ```

## Basic Usage

### Add a YubiKey

```bash
curl -X POST http://localhost:8002/api/v1/keys \
  -H "Content-Type: application/json" \
  -d '{
    "key_id": "cccccccccccc",
    "aes_key": "MTIzNDU2Nzg5MDEyMzQ1Ng==",
    "description": "John Doe YubiKey"
  }'
```

### Validate OTP

```bash
curl -X POST http://localhost:8002/api/v1/decrypt \
  -H "Content-Type: application/json" \
  -d '{"otp": "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv"}'
```

### List Keys

```bash
curl http://localhost:8002/api/v1/keys
```

