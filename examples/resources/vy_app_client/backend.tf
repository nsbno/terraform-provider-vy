resource "vy_app_client" "backend_application" {
  name = "infrademo-backend.vydev.io"
  type = "backend"

  scopes = [
	"https://infrademo.vydev.io/demo/read"  # Refers to the resource server defined above
  ]
}
