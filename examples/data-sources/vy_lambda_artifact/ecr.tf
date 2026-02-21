# Get information about an ECR Image based on GitHub repository name
data "vy_lambda_artifact" "this_in_ecr" {
  github_repository_name = "infrademo-demo-app"
  ecr_repository_name    = "infrademo-demo-repo"
}

# Use the ECR artifact in a Lambda module
module "lambda_in_ecr" {
  source = "github.com/nsbno/terraform-aws-lambda?ref=1.0.0"

  service_name  = "my-function"
  artifact_type = "ecr"
  artifact      = data.vy_lambda_artifact.this_in_ecr
}
