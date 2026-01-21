resource "rabbitmq_permissions" "test" {
  user      = "guest"
  vhost     = "/"
  configure = ".*"
  write     = ".*"
  read      = ".*"
}
