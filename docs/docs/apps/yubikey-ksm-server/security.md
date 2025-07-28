# Security

## How YubiKey AES Keys Are Stored

YubiKey AES keys are encrypted before storage using AES-256-GCM encryption.

### Encryption Process

1. **Master key** is hashed with SHA-256
2. **Row key** is derived using HKDF with YubiKey ID as info parameter
3. **AES key** is encrypted with AES-GCM using the row key
4. **Result** is base64-encoded and stored in database

### Storage Format

```text
Database: key_id="cccccccccccc", encrypted_key="base64(nonce+ciphertext)"
```

### Security Properties

- Each YubiKey gets unique encryption key
- Random nonce prevents identical ciphertext
- GCM provides authentication (tamper detection)
- Only 16-byte AES keys are encrypted, metadata is plaintext

### Master Key

Generate with:

```bash
openssl rand -base64 48
```

Master key compromise requires re-encryption of all stored keys.

## Encryption Flow

```mermaid
flowchart TD
    A[API Request: AES Key] --> B{Input Format}
    B -->|32 hex chars| C[Parse as Hex]
    B -->|Other format| D[Parse as Base64]
    C --> E[16-byte AES Key]
    D --> E
    E --> F[Master Key SHA-256]
    F --> G[HKDF: Derive Row Key]
    G --> H[Generate Random Nonce]
    H --> I[AES-256-GCM Encrypt]
    I --> J[Base64 Encode]
    J --> K[Store in Database]
    
    L[OTP Validation] --> M[Extract YubiKey ID]
    M --> N[Query Database]
    N --> O[Base64 Decode]
    O --> P[Split Nonce/Ciphertext]
    P --> Q[HKDF: Derive Same Row Key]
    Q --> R[AES-256-GCM Decrypt]
    R --> S[16-byte AES Key]
    S --> T[Decrypt YubiKey OTP]
```

## Standards Compliance

- **FIPS 140-2**: Uses FIPS-approved algorithms (AES-256, SHA-256, HKDF)
- **NIST SP 800-108**: HKDF key derivation follows NIST recommendations
- **RFC 5869**: HKDF implementation per RFC specification
- **RFC 5116**: AES-GCM authenticated encryption per RFC standard
