# Tsuribari

A secure webhook processing service that receives, validates, stores,
and transforms webhooks into workflow messages, stored in CouchDB and
published via AMQP for downstream processing.

## Features

- IP filtering and HMAC signature validation
- persistent storage in CouchDB with deduplication
- converts gitHub webhooks to workflow messages
- publishes workflows to RabbitMQ for downstream processing
- supports multiple organizations with separate secrets

## Architecture

- WebHook → IP Filter → HMAC Validation → CouchDB → Workflow → RabbitMQ

## Prerequisites

- CouchDB
- LavinMQ or RabbitMQ

## Build

```
$ make dist
```

## Usage

- create a `config.yml` file in `/usr/local/etc/tsuribari/`
- add the required HMAC secrets per organisation
- set the trusted IPs that can send webhooks
- run the server with `tsuribari`

```yaml
server:
  host: "127.0.0.1"
  port: "4003"

couchdb:
  url: "http://admin:passwd@127.0.0.1:5984/"
  database: "your_db"

rabbitmq:
  url: "amqp://guest:guest@localhost:5672/"
  exchange: "your.topic"
  queue: "your.queue"

security:
  trusted_ips:
    - "123.45.67.89"
  secrets:
    demo: "123456789abcdef0123456789abcdef0"
```

## Webhook Endpoints and Usage

### Basic Webhook
```
POST /webhooks/{organisation}
```

### Pipeline-Specific Webhook
```
POST /webhooks/{organisation}/{pipeline}
```

### Required Headers

Webhooks must include one of these signature headers:
- `X-Hub-Signature: sha1=<hmac-signature>`
- `X-Koan-Signature: sha256=<hmac-signature>`

### Example Request

```shell
curl -X POST \
  http://127.0.0.1:4003/webhooks/demo \
  -H "Content-Type: application/json" \::1/128"
  -H "X-Hub-Signature: sha1=<calculated-hmac>" \
  -d '{
    "repository": {
      "ssh_url": "git@github.com:demo/repo.git",
      "owner": {
        "login": "demo"
      }
    },
    "head_commit": {
      "id": "abc123def456"
    }
  }'
```

### HMAC Signature Calculation

Generate the HMAC-SHA1 signature using your organization's secret:

```shell
# Example using openssl
echo -n '{"repository":{"ssh_url":"git@github.com:demo/repo.git"}}' | \
  openssl dgst -sha1 -hmac "123456789abcdef0123456789abcdef0"
```

## API Responses

### Success Response
```json
{
  "message": "you have achieved enlightenment"
}
```

### Error Responses

#### Invalid IP
```json
{
  "error": "forbidden"
}
```
Response headers: `X-Capnhook: invalid source ip`

#### Invalid HMAC
```json
{
  "error": "invalid hmac"
}
```
Response headers: `X-Capnhook: invalid hmac`

#### Missing Secret
```json
{
  "error": "forbidden"
}
```
Response headers: `X-Capnhook: no secret found`

## Health Check

```
GET /healthz
```

Returns HTTP 200 OK when the service is running.

## Data Flow

1. **Webhook Reception**: Incoming webhook is received at the endpoint
2. **IP Validation**: Source IP is checked against trusted IP list
3. **HMAC Validation**: Webhook signature is verified using organization secret
4. **Storage**: Webhook is stored in CouchDB with SHA1-based deduplication
5. **Transformation**: GitHub webhook is transformed into workflow format
6. **Publishing**: Workflow message is published to RabbitMQ queue

## Workflow Message Format

Transformed workflows have this structure:

```json
{
  "id": "document-sha1-hash",
  "ref": "commit-id",
  "url": "repository-ssh-url",
  "org": "organization-name",
  "cache": "repository-url-sha256-hash",
  "utc": "2023-01-01T12:00:00Z"
}
```

## Security Features

- IP Filtering: Only trusted IPs can send webhooks
- HMAC Validation: Cryptographic signature verification
- Organization Isolation: Separate secrets per organization
- Deduplication: Prevents duplicate webhook processing

## Monitoring

The application logs important events including:
- Server startup
- Configuration loading
- Database/queue connection status
- Webhook processing results

## Development

### Project Structure
```
tsuribari/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── handlers/       # HTTP request handlers
│   ├── middleware/     # Security middleware
│   ├── models/         # Data structures
│   ├── queue/          # RabbitMQ integration
│   └── storage/        # CouchDB integration
├── config.yml.example  # Configuration file
└── go.mod              # Go module definition
```

