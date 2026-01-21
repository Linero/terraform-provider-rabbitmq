resource "rabbitmq_user" "test" {
  name                = "test"
  password_wo         = "test"
  password_wo_version = "1"
  tags = [
    "administrator"
  ]
}
