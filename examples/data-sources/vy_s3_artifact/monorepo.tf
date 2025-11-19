# For monorepos, you can specify a working directory within the repository where the lambda code is stored
data "vy_s3_artifact" "user-service" {
  github_repository_name = "infrademo-demo-app"
  working_directory      = "services/user-service"
}

# Use the S3 artifact in a Lambda module
module "lambda" {
  source = "github.com/nsbno/terraform-aws-lambda?ref=x.y.z"

  service_name  = "my-function"
  artifact_type = "s3"
  artifact      = data.vy_s3_artifact.user-service
}