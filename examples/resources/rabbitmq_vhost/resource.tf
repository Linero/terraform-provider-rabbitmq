resource "rabbitmq_vhost" "test" {
  name               = "test"
  description        = "My test vhost"
  default_queue_type = "quorum"
  tracing            = true
  tags               = ["tag1", "tag2"]
}