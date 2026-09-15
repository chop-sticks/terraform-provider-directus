# NOTE: This file is for HashiCorp specific licensing automation and can be deleted after creating a new repo with this template.
schema_version = 1

project {
  license        = "MIT"
  copyright_year = 2026

  header_ignore = [
    # jetbrains ides configs
    ".idea/**/*.xml",

    # examples used within documentation (prose)
    # "examples/**",

    # golangci-lint tooling configuration
    ".golangci.yml",

    # GoReleaser tooling configuration
    ".goreleaser.yml",
  ]
}