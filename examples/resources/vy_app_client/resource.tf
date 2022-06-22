resource "vy_app_client" "test" {
  name = "app_client_basic.acceptancetest.io"
  type = "backend"
  scopes = [
    "my.cool.service.vydev.io/read"
  ]
}
