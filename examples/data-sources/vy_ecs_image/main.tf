data "vy_ecs_image" "this" {
  github_repository_name = "infrademo-demo-app"
}

# Use the version in an ECS task definition
module "task" {
  source = "github.com/nsbno/terraform-aws-ecs-service?ref=x.y.z"

  application_container = {
    name  = "backend"
    image = data.vy_ecs_image.this
  }
}
