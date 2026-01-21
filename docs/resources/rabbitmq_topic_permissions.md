---
page_title: "rabbitmq_topic_permissions Resource - rabbitmq"
description: |-
  Resource to create and manage RabbitMQ topic permissions.
---

# rabbitmq_topic_permissions (Resource)

Resource to create and manage RabbitMQ topic permissions for a user.

## Example Usage

```terraform
resource "rabbitmq_topic_permissions" "test" {
  user     = "guest"
  vhost    = "/"
  exchange = "amq.topic"
  write    = ".*"
  read     = ".*"
}
```

## Schema

### Required

- `user` (String) The user to grant permissions to.
- `exchange` (String) The exchange to apply permissions to.
- `write` (String) The write permissions.
- `read` (String) The read permissions.

### Optional

- `vhost` (String) The vhost to grant permissions for. Defaults to `/`.

### Read-Only

- `id` (String) The ID of this resource.

## Import

`rabbitmq_topic_permissions` can be imported using the user name, vhost and exchange, e.g.

```
$ terraform import rabbitmq_topic_permissions.test guest@/@amq.topic
```
