# Get information about an ECR Image based on GitHub repository name
data "vy_lambda_artifact" "this" {
  github_repository_name = "infrademo-demo-app"
  ecr_repository_name    = "infrademo-demo-repo"
}

# Use the ECR artifact in a Lambda module
module "lambda" {
  source = "github.com/nsbno/terraform-aws-lambda?ref=x.y.z"

  service_name  = "my-function"
  artifact_type = "ecr"
  artifact      = data.vy_lambda_artifact.this
}
