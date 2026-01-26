---
page_title: "rabbitmq_vhost"
---

# rabbitmq_vhost

The `rabbitmq_vhost` resource creates and manages a vhost.

## Example Usage

```hcl
resource "rabbitmq_vhost" "test" {
  name = "test"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the vhost.

## Import

Vhosts can be imported using the `name`, e.g.

```
terraform import rabbitmq_vhost.test test
```
