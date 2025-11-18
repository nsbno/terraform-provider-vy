required_providers {
  vy = {
	source  = "nsbno/vy"
	version = ">= 0.5.0, < 1.0.0"
  }
}

provider "vy" {
  environment = "prod"
}
