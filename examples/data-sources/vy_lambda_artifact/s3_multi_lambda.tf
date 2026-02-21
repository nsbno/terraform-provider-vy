data "vy_lambda_artifact" "lambda_1" {
  github_repository_name = "infrademo-demo-app"
  // The sub-path under infrademo-demo-app in the S3 artifact bucket
  path = "apps/lambda-1"
}

data "vy_lambda_artifact" "lambda_2" {
  github_repository_name = "infrademo-demo-app"
  // Another sub-path under infrademo-demo-app
  path = "apps/lambda-2"
}

module "lambda_function_1" {
  source = "github.com/nsbno/terraform-aws-lambda?ref=1.0.0"

  service_name   = "my-function"
  component_name = "lambda-1"

  artifact_type = "s3"
  artifact      = data.vy_lambda_artifact.lambda_1
}

module "lambda_function_2" {
  source = "github.com/nsbno/terraform-aws-lambda?ref=1.0.0"

  service_name   = "my-function"
  component_name = "lambda-2"

  artifact_type = "s3"
  artifact      = data.vy_lambda_artifact.lambda_2
}