# Get information about an S3 artifact based on GitHub repository name
data "vy_lambda_artifact" "app_in_s3" {
  github_repository_name = "infrademo-demo-app"
}

# Use the S3 artifact in a Lambda module
module "lambda_in_s3" {
  source = "github.com/nsbno/terraform-aws-lambda?ref=1.0.0"

  service_name  = "my-function"
  artifact_type = "s3"
  artifact      = data.vy_lambda_artifact.app_in_s3
}
