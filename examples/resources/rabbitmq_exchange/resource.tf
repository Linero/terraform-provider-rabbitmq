resource "rabbitmq_exchange" "test" {
  name  = "test-exchange"
  vhost = "/"
  settings {
    type        = "direct"
    durable     = false
    auto_delete = false
  }
}
