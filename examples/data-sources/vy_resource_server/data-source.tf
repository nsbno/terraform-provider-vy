# Look up an existing resource server owned by another module or team
data "vy_resource_server" "reiserad_backend" {
  identifier = "https://services.trafficcontrol.vydev.io/reiserad-backend"
}
