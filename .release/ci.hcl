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
event "upload-dev" {

  action "upload-dev" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "upload-dev"
    depends      = ["build"]
    config       = ""
  }

  depends = ["build"]

  notification {
    on = "fail"
  }
}
event "security-scan-binaries" {

  action "security-scan-binaries" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "security-scan-binaries"
    depends      = null
    config       = "security-scan.hcl"
  }

  depends = ["upload-dev"]

  notification {
    on = "fail"
  }
}
event "sign" {

  action "sign" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "sign"
    depends      = null
    config       = ""
  }

  depends = ["security-scan-binaries"]

  notification {
    on = "fail"
  }
}
event "verify" {

  action "verify" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "verify"
    depends      = null
    config       = ""
  }

  depends = ["sign"]

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