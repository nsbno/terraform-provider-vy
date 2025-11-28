data "vy_ecs_image" "this" {
  github_repository_name = "infrademo-demo-app"

  # To override the ECR repository name, specify it here
  ecr_repository_name = "petstore-lambda"
}

module "lambda" {
  source = "github.com/nsbno/terraform-aws-ecs-service?ref=x.y.z"

  application_container = {
    name  = "user-service"
    image = data.vy_ecs_image.this
  }
}
