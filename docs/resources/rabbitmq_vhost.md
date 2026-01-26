---
page_title: "rabbitmq_vhost Resource - rabbitmq"
description: |-
  Resource to create and manage a RabbitMQ vhost.
---

# rabbitmq_vhost (Resource)

Resource to create and manage a RabbitMQ vhost.

## Example Usage

```terraform
resource "rabbitmq_vhost" "test" {
  name               = "test"
  description        = "My test vhost"
  default_queue_type = "quorum"
  tracing            = true
  tags               = ["tag1", "tag2"]
}
```

## Schema

### Required

- `name` (String) The name of the vhost.

### Optional

- `description` (String) The description of the vhost.
- `default_queue_type` (String) The default queue type for the vhost.
- `tags` (List of String) Tags associated with the vhost.
- `tracing` (Boolean) The tracing setting for the vhost.

### Read-Only

- `id` (String) The ID of this resource.

## Import

`rabbitmq_vhost` can be imported using the name, e.g.

```
$ terraform import rabbitmq_vhost.test test
```
