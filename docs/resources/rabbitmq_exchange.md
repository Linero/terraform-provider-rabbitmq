---
page_title: "rabbitmq_exchange Resource - rabbitmq"
description: |-
  Resource to create and manage RabbitMQ exchanges.
---

# rabbitmq_exchange (Resource)

Resource to create and manage RabbitMQ exchanges.

## Example Usage

```terraform
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

- `name` (String) The name of the exchange.
- `settings` (Block List, Min: 1, Max: 1) A nested block to configure the exchange settings. (see below for nested schema)

### Optional

- `vhost` (String) The vhost to create the exchange in. Defaults to `/`.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--settings"></a>
### Nested Schema for `settings`

Required:

- `type` (String) The type of the exchange.

Optional:

- `auto_delete` (Boolean) Whether the exchange will be automatically deleted when no longer in use.
- `durable` (Boolean) Whether the exchange is durable.
- `arguments` (Map of String) Additional arguments for the exchange.

## Import

`rabbitmq_exchange` can be imported using the name and vhost, e.g.

```
$ terraform import rabbitmq_exchange.test test-exchange@/
```
