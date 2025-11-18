# Enroll an AWS account as an environment account in Vy's deployment system
resource "vy_environment_account" "this" {
  owner_account_id = "123456789012"
}
