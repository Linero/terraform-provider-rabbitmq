# Rabbitmq Provider

The Rabbitmq provider provides resources to interact with a Rabbitmq server, specifically for managing users, permissions and topic permissions.

## Example Usage

```terraform
provider "rabbitmq" {
  address    = "http://localhost:15672"
  username   = "guest"
  password   = "guest"
}

resource "rabbitmq_user" "test" {
  name                = "test"
  password_wo         = "test"
  password_wo_version = "1"
  tags = [
    "administrator"
  ]
}

resource "rabbitmq_permissions" "test" {
  user      = "test"
  vhost     = "/"
  configure = ".*"
  write     = ".*"
  read      = ".*"
}

resource "rabbitmq_topic_permissions" "test" {
  user     = "test"
  vhost    = "/"
  exchange = "amq.topic"
    write     = ".*"
    read      = ".*"
  }
  
  resource "rabbitmq_exchange" "test" {
    name  = "test-exchange"
    vhost = "/"
    settings {
      type        = "direct"
      durable     = false
      auto_delete = false
    }
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
