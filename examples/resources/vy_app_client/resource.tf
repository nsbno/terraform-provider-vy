resource "vy_app_client" "backend_application" {
  name = "app_client_basic.acceptancetest.io"
  type = "backend"

  scopes = [
    "my.cool.service.vydev.io/read",
    "my.cool.service.vydev.io/delete",
  ]
}

resource "vy_app_client" "frontend_application" {
  name = "app_client_basic.acceptancetest.io"
  type = "frontend"

  scopes = [
    "my.cool.service.vydev.io/read",
    "my.cool.service.vydev.io/write",
  ]

  callback_urls = ["https://example.com/callback"]
  logout_urls   = ["https://example.com/logout"]
}
