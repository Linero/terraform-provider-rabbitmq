resource "rabbitmq_topic_permissions" "test" {
  user     = "test"
  vhost    = "/"
  exchange = "amq.topic"
  write    = ".*"
  read     = ".*"
}
