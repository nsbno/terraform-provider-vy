# Get information about an ECR Image based on ECR Repository Name
data "vy_ecr_image" "this" {
  ecr_repository_name = "infrademo-demo-app"
}

# Use the version in an ECS task definition
module "task" {
  source = "github.com/nsbno/terraform-aws-ecs-service?ref=x.y.z"

  application_container = {
	name  = "backend"
	image = data.vy_ecr_image.this
  }
}
