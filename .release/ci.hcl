schema = "2"

project "vault-ssh-helper" {
  team = "vault"

  slack {
    notification_channel = "C03RXFX5M4L"
  }

  github {
    organization     = "hashicorp"
    repository       = "vault-ssh-helper"
    release_branches = ["main"]
  }
}

event "merge" {
}
event "build" {

  action "build" {
    organization = "hashicorp"
    repository   = "vault-ssh-helper"
    workflow     = "build"
    depends      = null
    config       = ""
  }

  depends = ["merge"]
}
event "prepare" {

  action "prepare" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "prepare"
    depends      = ["build"]
    config       = ""
  }

  depends = ["build"]

  notification {
    on = "fail"
  }
}
event "trigger-staging" {
}
event "promote-staging" {

  action "promote-staging" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "promote-staging"
    depends      = null
    config       = "release-metadata.hcl"
  }

  depends = ["trigger-staging"]

  notification {
    on = "always"
  }

  promotion-events {
  }
}
event "trigger-production" {
}
event "promote-production" {

  action "promote-production" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "promote-production"
    depends      = null
    config       = ""
  }

  depends = ["trigger-production"]

  notification {
    on = "always"
  }

  promotion-events {
  }
}