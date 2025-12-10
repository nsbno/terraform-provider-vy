data "aws_caller_identity" "current" {}

resource "vy_app_client" "client" {
  name = "${data.aws_caller_identity.current.account_id}-infrademo"
  type = "frontend"

  callback_urls   = [
	"http://localhost:3000/auth/callback",
	"https://petstore.infrademo.vydev.io/auth/callback",  # Example
  ]
  logout_urls = [
	"http://localhost:3000/logout",
	"https://petstore.infrademo.vydev.io/logout",  # Example
  ]

  scopes = [
	"email",
	"openid",
	"profile",
  ]
}