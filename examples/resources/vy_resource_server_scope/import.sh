# Scopes can be imported using the format: <resource_server_identifier>,<scope_name>
# <scope_name> is the full scope name as stored remotely, i.e. including any "namespace/" prefix.
terraform import vy_resource_server_scope.read ruteplan.vydev.io,fillrate/read
