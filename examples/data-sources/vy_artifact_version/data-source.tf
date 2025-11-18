# Get the latest version of a Lambda artifact from S3
data "vy_artifact_version" "lambda" {
  application = "my-lambda-function"  # Lambda Name
}

# Use the version in a Lambda module
module "lambda" {
  source = "github.com/nsbno/terraform-aws-lambda?ref=x.y.z"

  name          = "my-function"
  artifact_type = "s3"
  artifact      = data.vy_artifact_version.lambda
}

# Get the latest version of a container image from ECR
data "vy_artifact_version" "server" {
  application = "my-backend-service"  # ECR Repository name
}

# Use the version in an ECS task definition
module "task" {
  source = "github.com/nsbno/terraform-aws-ecs-service?ref=x.y.z"

  application_container = {
    name  = "backend"
    image = "${data.vy_artifact_version.server.store}/${data.vy_artifact_version.server.path}@${data.vy_artifact_version.server.version}"
  }
}
