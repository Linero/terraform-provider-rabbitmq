---
page_title: "rabbitmq_user Resource - rabbitmq"
description: |-
  Resource to create and manage a RabbitMQ user.
---

# rabbitmq_user (Resource)

Resource to create and manage a RabbitMQ user.

## Example Usage

```terraform
resource "rabbitmq_user" "test" {
  name                = "test"
  password_wo         = "test"
  password_wo_version = "1"
  tags = [
    "administrator"
  ]
}
```

## Schema

### Required

- `name` (String) Name of the user.
- `password_wo` (String, Sensitive, Write-Only) Write-only password for the user.
- `password_wo_version` (String) Version string for password. Changing this value forces a password update even if password_wo hasn't changed in the configuration. Use this to rotate passwords.

### Optional

- `tags` (List of String) Tags for the user.

### Read-Only

- `id` (String) The ID of this resource.

## Import

`rabbitmq_user` can be imported using the user name, e.g.

```
$ terraform import rabbitmq_user.test test
```