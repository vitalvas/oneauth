# Configuration

## Basic Configuration

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

## Server

| Option | Default | Description |
|--------|---------|-------------|
| `address` | `localhost:8002` | Server bind address |

## Database

### SQLite (Default)

```yaml
database:
  type: sqlite
  sqlite:
    path: /var/lib/oneauth/yubikey_ksm.db
    journal_mode: WAL      # Optional
    synchronous: NORMAL    # Optional
```

### PostgreSQL

```yaml
database:
  type: postgres
  postgres:
    url: "postgres://ksm_user:secret_password@localhost:5432/yubikey_ksm?sslmode=disable"
    max_connections: 25
    connection_timeout: 30s
```

## Security

| Option | Required | Description |
|--------|----------|-------------|
| `master_key` | Yes | Master encryption key (32+ characters) |

## Logging

| Option | Default | Options |
|--------|---------|---------|
| `level` | `info` | `debug`, `info`, `warn`, `error` |
| `format` | `json` | `json`, `text` |

## Environment Variables

Override config with `ONEAUTH_KSM_` prefix:

```bash
export ONEAUTH_KSM_SERVER_ADDRESS="0.0.0.0:8002"
export ONEAUTH_KSM_SECURITY_MASTER_KEY="$(openssl rand -base64 48)"
export ONEAUTH_KSM_DATABASE_TYPE="postgres"
```
