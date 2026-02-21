terraform {
  required_providers {
    vy = {
      source  = "nsbno/vy"
      version = "1.2.0"
    }
  }
}

provider "vy" {
  environment = "test"
}