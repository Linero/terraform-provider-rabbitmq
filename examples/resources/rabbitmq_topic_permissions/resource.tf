resource "rabbitmq_topic_permissions" "test" {
  user     = "guest"
  vhost    = "/"
  exchange = "amq.topic"
  write    = ".*"
  read     = ".*"
}
