terraform {
  required_providers {
    imgix = {
      source = "journey.travel/terraform-providers/imgix"
      version = "1.0.0"
    }
  }
}

provider "imgix" {}