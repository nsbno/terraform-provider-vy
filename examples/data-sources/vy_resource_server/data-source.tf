# Look up an existing resource server owned by another module or team
data "vy_resource_server" "bounded_context" {
  identifier = "ruteplan.vydev.io"
}
