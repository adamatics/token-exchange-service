# Token Exchange Service

A microservice for exchanging Azure AD tokens using OAuth 2.0 On-Behalf-Of flow.

## Features

-   Token exchange using Azure AD's On-Behalf-Of flow
-   Interactive API documentation (Swagger UI)
-   Docker support
-   Environment-based configuration

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

## API Documentation

-   Interactive documentation: `GET /`
-   OpenAPI spec: `GET /swagger.json` or `GET /swagger.yaml`

## Environment Variables

| Variable      | Required | Default |
| ------------- | -------- | ------- |
| CLIENT_ID     | Yes      | -       |
| CLIENT_SECRET | Yes      | -       |
| TENANT_ID     | Yes      | -       |
| DEFAULT_SCOPE | No       | -       |
| PORT          | No       | 9000    |

## Usage

### Curl

Exchange a token:

```bash
curl -X POST http://localhost:9000/ \
  -H "Content-Type: application/json" \
  -d '{
    "adalab_token": "adalab_access_token",
    "scopes": ["https://graph.microsoft.com/.default"]
  }'
```

### AdaLab JupyterLab

```python
"""
Example script to acquire a scoped Azure AD access token on behalf of a user
through a multi-step token exchange. Designed to be run in a Jupyter cell.
"""
import requests
import json
from requests.exceptions import HTTPError, RequestException

# These functions are part of an internal library for auth and config.
from adalib_auth.keycloak import get_client_token, jupyterhub_authentication
from adalib_auth.config import get_config

# --- Configuration (CHANGE ME) ---
TOKEN_EXCHANGE_APP_URL = "https://<YOUR_ADALAB_URL_HERE>/apps/token-exchange"
TARGET_API_SCOPE = "api://<YOUR_SCOPE_HERE>"

try:
    # Step 1: Get initial service and user tokens.
    adalib_config = get_config()
    platform_token = get_client_token("adaboard")["access_token"]
    jupyterhub_token = jupyterhub_authentication("jupyterhub")["access_token"]

    # Step 2: Broker the user's token for an Azure AD token.
    keycloak_base_url = f"{adalib_config.SERVICES['keycloak']['url']}/auth/"
    azure_broker_url = f"{keycloak_base_url}realms/{adalib_config.KEYCLOAK_REALM}/broker/azure/token"

    azure_response = requests.get(
        azure_broker_url,
        headers={"Authorization": f"Bearer {jupyterhub_token}"},
    )
    azure_response.raise_for_status()
    azure_refresh_token = azure_response.json()["refresh_token"]

    # Step 3: Use the exchange service to refresh the Azure token.
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {platform_token}"
    }
    refresh_response = requests.post(
        f"{TOKEN_EXCHANGE_APP_URL}/refresh",
        headers=headers,
        json={"refresh_token": azure_refresh_token}
    )
    refresh_response.raise_for_status()
    refreshed_access_token = refresh_response.json()["access_token"]

    # Step 4: Perform the final On-Behalf-Of exchange for the target scope.
    obo_response = requests.post(
        f"{TOKEN_EXCHANGE_APP_URL}/",
        headers=headers,
        json={"adalab_token": refreshed_access_token, "scopes": [TARGET_API_SCOPE]}
    )
    obo_response.raise_for_status()
    final_tokens = obo_response.json()

    # --- Success ---
    print("Token exchange successful. Final token details:")
    print(json.dumps(final_tokens, indent=2))
    # The final access token can now be used for subsequent API calls, for example:
    # final_access_token = final_tokens["access_token"]

except (RequestException, HTTPError, KeyError, Exception) as err:
    print(f"An error occurred during token exchange: {err}")
```
