resource "vy_resource_server" "this" {
  identifier = "service.vydev.io"
  name       = "my service"
  scopes = [
    {
      name        = "read"
      description = "used for reading"
    }
  ]
}