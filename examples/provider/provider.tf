# Copyright IBM Corp. 2026
# SPDX-License-Identifier: MIT

terraform {
  required_providers {
    directus = {
      source = "chop-sticks/directus"
    }
  }
}

provider "directus" {
  url   = "http://example.localhost:8055"
  token = "secret_token"
}

data "directus_collection" "builtin_users" {
  collection = "directus_users"
}

output "users" {
  value = data.directus_collection.builtin_users
}