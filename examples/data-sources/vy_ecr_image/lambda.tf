# Get information about an ECR Image based on ECR Repository Name
data "vy_ecr_image" "this" {
  ecr_repository_name = "infrademo-demo-app"
}

# Use the image in a Lambda module
module "lambda" {
  source = "github.com/nsbno/terraform-aws-lambda?ref=x.y.z"

  service_name  = "my-function"
  artifact_type = "ecr"
  image         = data.vy_ecr_image.this
}
