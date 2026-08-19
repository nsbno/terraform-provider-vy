data "vy_resource_server" "bounded_context" {
  identifier = "ruteplan.vydev.io"
}

resource "vy_resource_server_scope" "read" {
  resource_server = data.vy_resource_server.bounded_context.identifier

  namespace   = "fillrate" # The name of the microservice/domain. Optional, but helps with namespacing
  name        = "read"
  description = "Used for reading"
}
