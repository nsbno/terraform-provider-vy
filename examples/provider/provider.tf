required_providers {
  vy = {
	source  = "nsbno/vy"
	version = ">= 1.0.0, < 2.0.0"
  }
}

provider "vy" {
  environment = "prod"
}
