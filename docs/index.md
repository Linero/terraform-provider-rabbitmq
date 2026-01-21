---
page_title: "Rabbitmq Provider"
description: |-
  The rabbitmq provider allows managing rabbitmq users.
---

# Rabbitmq Provider

The Rabbitmq provider provides resources to interact with a Rabbitmq server, specifically for managing users.

## Example Usage

```terraform
provider "rabbitmq" {
  address    = "http://localhost:15672"
  username   = "guest"
  password   = "guest"
}
```

## Schema

### Required

- `address` (String) The address of the Rabbitmq server (e.g., `http://localhost:15672`).
- `username` (String, Sensitive) The username for the Rabbitmq user.
- `password` (String, Sensitive) The password for the Rabbitmq user.

### Optional

- `insecure` (Boolean) Trust self-signed certificates.
- `cacert_file` (String) Path to the CA certificate file.
- `clientcert_file` (String) Path to the client certificate file.
- `clientkey_file` (String) Path to the client key file.
- `proxy` (String) Proxy URL to use for requests.