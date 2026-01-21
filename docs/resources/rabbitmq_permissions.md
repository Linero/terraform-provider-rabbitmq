---
page_title: "rabbitmq_permissions Resource - rabbitmq"
description: |-
  Resource to create and manage RabbitMQ permissions.
---

# rabbitmq_permissions (Resource)

Resource to create and manage RabbitMQ permissions for a user.

## Example Usage

```terraform
resource "rabbitmq_permissions" "test" {
  user      = "test"
  vhost     = "/"
  configure = ".*"
  write     = ".*"
  read      = ".*"
}
```

## Schema

### Required

- `user` (String) The user to grant permissions to.
- `configure` (String) The configure permissions.
- `write` (String) The write permissions.
- `read` (String) The read permissions.

### Optional

- `vhost` (String) The vhost to grant permissions for. Defaults to `/`.

### Read-Only

- `id` (String) The ID of this resource.

## Import

`rabbitmq_permissions` can be imported using the user name and vhost, e.g.

```
$ terraform import rabbitmq_permissions.test guest@/
```
