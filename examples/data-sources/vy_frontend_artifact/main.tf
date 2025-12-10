terraform {
  required_providers {
    staticfiledeploy = {
      source  = "nsbno/static-file-deploy"
      version = ">= 1.0.0, < 2.0.0"
    }
  }
}
# Get information about a frontend artifact based on GitHub repository name
data "vy_frontend_artifact" "this" {
  github_repository_name = "infrademo-static-website"
}

# Use together with: https://github.com/nsbno/terraform-provider-static-file-deploy
resource "staticfiledeploy_deployment" "this" {
  source         = data.vy_frontend_artifact.this.s3_source_path
  source_version = data.vy_frontend_artifact.this.s3_object_version
  target         = aws_s3_bucket.frontend.bucket
}

resource "aws_s3_bucket" "frontend" {
  bucket = "infrademo-static-website-frontend"
}
