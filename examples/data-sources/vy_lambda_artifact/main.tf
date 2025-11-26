# Get information about an S3 artifact based on GitHub repository name
data "vy_lambda_artifact" "this" {
  github_repository_name = "infrademo-demo-app"
}

# Use the S3 artifact in a Lambda module
module "lambda" {
  source = "github.com/nsbno/terraform-aws-lambda?ref=x.y.z"

  service_name  = "my-function"
  artifact_type = "s3"
  artifact      = data.vy_lambda_artifact.this
}
