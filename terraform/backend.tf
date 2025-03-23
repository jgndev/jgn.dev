terraform {
  cloud {
    organization = "stratocraft-dev"

    workspaces {
      name = "stratocraft-dev"
    }
  }
}