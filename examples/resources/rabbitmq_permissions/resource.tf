resource "rabbitmq_permissions" "test" {
  user      = "test"
  vhost     = "/"
  configure = ".*"
  write     = ".*"
  read      = ".*"
}
