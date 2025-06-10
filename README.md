# Token Exchange Service

A microservice for exchanging Azure AD tokens using OAuth 2.0 On-Behalf-Of flow.

## Features
- Token exchange using Azure AD's On-Behalf-Of flow
- Interactive API documentation (Swagger UI)
- Docker support
- Environment-based configuration

## Quick Start

### Local Development
```bash
# Install dependencies and tools
go mod download
go install github.com/swaggo/swag/cmd/swag@latest

# Generate API docs
swag init

# Set environment variables
export CLIENT_ID=your_client_id
export CLIENT_SECRET=your_client_secret
export TENANT_ID=your_tenant_id

# Run
go run main.go
```

### Docker
```bash
docker run -p 9000:9000 \
  -e CLIENT_ID=your_client_id \
  -e CLIENT_SECRET=your_client_secret \
  -e TENANT_ID=your_tenant_id \
  token-exchange-service
```

## Usage

Exchange a token:
```bash
curl -X POST http://localhost:9000/ \
  -H "Content-Type: application/json" \
  -d '{
    "adalab_token": "adalab_access_token",
    "scopes": ["https://graph.microsoft.com/.default"]
  }'
```

## API Documentation
- Interactive documentation: `GET /` 
- OpenAPI spec: `GET /swagger.json` or `GET /swagger.yaml`

## Environment Variables

| Variable | Required | Default |
|----------|----------|---------|
| CLIENT_ID | Yes | - |
| CLIENT_SECRET | Yes | - |
| TENANT_ID | Yes | - |
| PORT | No | 9000 |
