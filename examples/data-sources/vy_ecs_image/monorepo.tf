data "vy_ecs_image" "user_service" {
  github_repository_name = "infrademo-demo-app"
  ecr_repository_name    = "infrademo-demo-repo"

  # Path to the directory within the monorepo where the service code is located
  working_directory = "services/user_service"
}

module "user_service_ecs" {
  source = "github.com/nsbno/terraform-aws-ecs-service?ref=x.y.z"

  application_container = {
    name  = "user-service"
    image = data.vy_ecs_image.user_service
  }
}

data "vy_ecs_image" "payment_service" {
  github_repository_name = "infrademo-demo-app"
  ecr_repository_name    = "infrademo-demo-repo"

  # Path to the directory within the monorepo where the service code is located
  working_directory = "services/payment_service"
}

module "payment_service_ecs" {
  source = "github.com/nsbno/terraform-aws-ecs-service?ref=x.y.z"

  application_container = {
    name  = "payment-service"
    image = data.vy_ecs_image.payment_service
  }
}
