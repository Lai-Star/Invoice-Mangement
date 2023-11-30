terraform {
  required_providers {
    vault = {
      source = "hashicorp/vault"
      version = "2.21.0"
    }
  }
}

provider "vault" {
  address = "https://vault.monetr.mini"
}