data "aws_caller_identity" "current" {}

# Scopes for this resource server are declared in the `scopes` list.
# Do not use `vy_resource_server_scope` to manage scopes separately.
resource "vy_resource_server" "this" {
  identifier = "https://services.myteam.vydev.io/incident"
  name       = "${data.aws_caller_identity.current.account_id}-myteam-incident"

  scopes = [
    {
      name        = "read"
      description = "Allows reading incidents"
    },
    {
      name        = "write"
      description = "Allows creating and updating incidents"
    }
  ]
}

# Alternatively, manage scopes independently with `vy_resource_server_scope`.
# Leave `scopes` unset here, and use `ignore_changes` as a safeguard against
# accidentally setting it later.
resource "vy_resource_server" "that" {
  identifier = "https://services.myteam.vydev.io/reports"
  name       = "${data.aws_caller_identity.current.account_id}-myteam-reports"

  lifecycle {
    ignore_changes = [scopes]
  }
}