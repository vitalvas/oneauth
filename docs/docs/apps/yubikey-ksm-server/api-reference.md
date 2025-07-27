# API Reference

## YubiKey Components

Each YubiKey has:

- **Key ID**: 12 modhex characters (e.g., `cccccccccccc`)
- **AES Key**: 16-byte secret key (base64 encoded)

Any YubiKey OTP contains the Key ID as the first 12 characters:

```text
ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv
^^^^^^^^^^^^
Key ID = cccccccccccc
```

Use `yubico-piv-tool -a status` to get key ID from YubiKey device.

## Key Formats

### Key ID

- **Length**: 12 characters
- **Characters**: `cbdefghijklnrtuv` (modhex only)
- **Example**: `cccccccccccc`

### AES Key

- **Length**: 16 bytes
- **Format**: Base64 encoded
- **Example**: `MTIzNDU2Nzg5MDEyMzQ1Ng==`

Use `xxd` and `base64` commands to convert hex AES keys to base64 format.

Key ID must be exactly 12 modhex characters (`cbdefghijklnrtuv`)

## Health Check

```bash
curl http://localhost:8002/api/v1/health
```

Response:

```json
{
  "status": "healthy",
  "database": {
    "status": "healthy"
  }
}
```

## Key Management

### Add Key

```bash
curl -X POST http://localhost:8002/api/v1/keys \
  -H "Content-Type: application/json" \
  -d '{
    "key_id": "cccccccccccc",
    "aes_key": "MTIzNDU2Nzg5MDEyMzQ1Ng==",
    "description": "John Doe YubiKey"
  }'
```

### List Keys

```bash
curl http://localhost:8002/api/v1/keys
```

Response:

```json
{
  "keys": [
    {
      "key_id": "cccccccccccc",
      "description": "John Doe YubiKey",
      "created_at": "2024-01-15T10:30:45Z"
    }
  ]
}
```

### Get Key

```bash
curl http://localhost:8002/api/v1/keys/cccccccccccc
```

### Delete Key

```bash
curl -X DELETE http://localhost:8002/api/v1/keys/cccccccccccc
```

## OTP Validation

### REST API

```bash
curl -X POST http://localhost:8002/api/v1/decrypt \
  -H "Content-Type: application/json" \
  -d '{"otp": "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv"}'
```

Success Response:

```json
{
  "status": "OK",
  "key_id": "cccccccccccc",
  "counter": 26,
  "timestamp_low": 35599,
  "timestamp_high": 15,
  "session_use": 3,
  "decrypted_at": "2024-01-15T10:30:45Z"
}
```

Error Response:

```json
{
  "status": "ERROR",
  "error_code": "KEY_NOT_FOUND",
  "message": "YubiKey not registered"
}
```

### KSM Protocol (Legacy)

```bash
curl "http://localhost:8002/wsapi/decrypt/?otp=ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv"
```

Success: `OK counter=001a low=8b0f high=0f use=03`
Error: `ERR Key not found`

## Error Codes

| Code | Description |
|------|-------------|
| `INVALID_OTP` | OTP format is invalid |
| `KEY_NOT_FOUND` | YubiKey not registered |
| `DECRYPTION_FAILED` | OTP decryption failed |

## Status Codes

| Code | Meaning |
|------|---------|
| `200` | Success |
| `201` | Created |
| `400` | Bad Request |
| `404` | Not Found |
| `500` | Server Error |
