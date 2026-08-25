data "vy_resource_server" "reiserad_backend" {
  identifier = "https://services.trafficcontrol.vydev.io/reiserad-backend"
}

resource "vy_resource_server_scope" "read" {
  resource_server = data.vy_resource_server.reiserad_backend.identifier

  # `namespace` is optional, e.g. your microservice or domain name.
  # It is combined with `name` as `namespace.name` (e.g. `reiserad.read`), allowing
  # teams sharing this resource server to avoid name collisions.
  namespace   = "reiserad"
  name        = "read"
  description = "Allows reading incidents"
}
