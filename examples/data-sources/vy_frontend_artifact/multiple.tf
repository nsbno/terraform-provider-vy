data "vy_frontend_artifact" "frontend_a" {
  github_repository_name = "infrademo-static-website"
  path                   = "apps/frontend-a"
}

data "vy_frontend_artifact" "frontend_b" {
  github_repository_name = "infrademo-static-website"
  path                   = "apps/frontend-b"
}

resource "staticfiledeploy_deployment" "frontend_a" {
  source         = data.vy_frontend_artifact.frontend_a.s3_source_path
  source_version = data.vy_frontend_artifact.frontend_a.s3_object_version
  target         = aws_s3_bucket.frontend_a.bucket
}

resource "staticfiledeploy_deployment" "frontend_b" {
  source         = data.vy_frontend_artifact.frontend_b.s3_source_path
  source_version = data.vy_frontend_artifact.frontend_b.s3_object_version
  target         = aws_s3_bucket.frontend_b.bucket
}
