# terraform-provider-directus

A [Terraform](https://www.terraform.io) provider for [Directus](https://directus.io),
built on the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
and the [`directus-client-go`](https://github.com/chop-sticks/directus-client-go) SDK.

It lets you manage a Directus instance's schema and configuration as code:
collections, fields, relations, roles, policies, permissions, flows, operations,
dashboards, panels, presets, translations, folders, users, and project settings.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- A reachable Directus instance and a static access token (or admin token)
- [Go](https://go.dev/dl/) >= 1.26 (only to build the provider from source)

## Using the provider

```hcl
terraform {
  required_providers {
    directus = {
      source = "chop-sticks/directus"
    }
  }
}

provider "directus" {
  url   = "https://directus.example.com" # or env DIRECTUS_URL
  token = var.directus_token             # or env DIRECTUS_TOKEN
}
```

### Provider configuration

| Argument | Type   | Required | Env              | Description                                   |
|----------|--------|----------|------------------|-----------------------------------------------|
| `url`    | string | yes\*    | `DIRECTUS_URL`   | Base URL of the Directus instance.            |
| `token`  | string | yes\*    | `DIRECTUS_TOKEN` | Static access token (sent as a bearer token). |

\* Each may be supplied via the argument **or** its environment variable. The
provider errors if neither is set.

```sh
export DIRECTUS_URL="https://directus.example.com"
export DIRECTUS_TOKEN="your-static-token"
```

## Example

```hcl
# A collection with a real database table.
resource "directus_collection" "articles" {
  collection = "articles"
  meta = {
    icon = "article"
    note = "Blog articles"
  }
}

# A field on that collection. `field` type-specific options are JSON strings.
resource "directus_field" "title" {
  collection = directus_collection.articles.collection
  field      = "title"
  type       = "string"
  meta = {
    interface = "input"
    required  = true
  }
  schema = {
    is_nullable = false
    max_length  = 255
  }
}

# Access control: a policy, a role that uses it, and a permission rule.
resource "directus_policy" "editor" {
  name       = "Editor"
  app_access = true
}

resource "directus_role" "editors" {
  name = "Editors"
  icon = "edit"
}

resource "directus_permission" "editors_read_articles" {
  policy     = directus_policy.editor.id
  collection = directus_collection.articles.collection
  action     = "read"
  fields     = ["*"]
}

# A user.
resource "directus_user" "alice" {
  email      = "alice@example.com"
  password   = var.alice_password # write-only; never read back from the API
  first_name = "Alice"
  role       = directus_role.editors.id
}

# Automation: a flow with a first operation.
resource "directus_flow" "notify" {
  name    = "Notify on publish"
  trigger = "event"
  status  = "active"
  options = jsonencode({ type = "action", scope = ["items.create"], collections = ["articles"] })
}

resource "directus_operation" "log" {
  flow       = directus_flow.notify.id
  key        = "log"
  type       = "log"
  position_x = 20
  position_y = 1
  options    = jsonencode({ message = "Article published" })
}

# Project settings (singleton).
resource "directus_settings" "this" {
  project_name    = "My Project"
  default_language = "en-US"
}
```

### Reading existing objects

```hcl
data "directus_collection" "existing" {
  collection = "articles"
}

data "directus_server_info" "this" {}

# `info` is the raw /server/info payload as a JSON string.
output "directus_server_info" {
  value = jsondecode(data.directus_server_info.this.info)
}
```

## Importing

```sh
# String-id resources (UUID):
terraform import directus_role.editors        <role-uuid>
terraform import directus_policy.editor       <policy-uuid>
terraform import directus_folder.docs         <folder-uuid>
terraform import directus_user.alice          <user-uuid>
terraform import directus_dashboard.ops       <dashboard-uuid>
terraform import directus_panel.chart         <panel-uuid>
terraform import directus_flow.notify         <flow-uuid>
terraform import directus_operation.log       <operation-uuid>
terraform import directus_translation.hello   <translation-uuid>

# Name-keyed:
terraform import directus_collection.articles articles

# Integer-keyed:
terraform import directus_preset.default_view 42
terraform import directus_permission.editors_read_articles 7

# Composite (collection/field):
terraform import directus_field.title     articles/title
terraform import directus_relation.author articles/author_id

# Singleton (any id works; the row is fixed):
terraform import directus_settings.this 1
```

## Resources and data sources

**Resources:** `directus_collection`, `directus_field`, `directus_relation`,
`directus_role`, `directus_policy`, `directus_permission`, `directus_folder`,
`directus_flow`, `directus_operation`, `directus_panel`, `directus_dashboard`,
`directus_preset`, `directus_translation`, `directus_settings`, `directus_user`.

**Data sources:** `directus_collection`, `directus_role`, `directus_policy`,
`directus_folder`, `directus_user`, `directus_settings`, `directus_server_info`.

Full attribute reference for each is generated under [`docs/`](./docs).

## Modeling notes

- **Free-form fields are JSON strings.** Open-ended Directus objects
  (`meta.options`, `field` `validation`/`options`, `permission` filters, flow/
  operation `options`, preset `filter`/`layout_query`, ...) are typed as
  normalized JSON — use `jsonencode(...)`. They compare semantically, so
  formatting and key order never cause drift.
- **`directus_collection` always creates a table.** A schema-less collection is
  a restricted "folder" in Directus; `Create` always sends a `schema` object.
- **Concurrent schema changes.** Creating several collections in a single applying
  can race Directus's schema cache. Serialize dependent schema mutations with
  `depends_on` when needed.
- **`directus_folder` has no update endpoint** — changing `name`/`parent`
  replaces the resource.
- **`directus_settings` is a singleton** — `terraform destroy` only drops it
  from state; it does not reset Directus settings.
- **`directus_user.password` is write-only** — the API never returns it, so it
  is kept from configuration and never triggers drift.

## Development

Common workflows are automated with [Task](https://taskfile.dev). Run `task` to
list them:

| Task                                    | Description                                                                                     |
|-----------------------------------------|-------------------------------------------------------------------------------------------------|
| `task build`                            | Compile all packages                                                                            |
| `task test`                             | Run the unit test suite                                                                         |
| `task testacc`                          | Run acceptance tests (`TF_ACC`) against a live Directus (needs `DIRECTUS_URL`/`DIRECTUS_TOKEN`) |
| `task testacc:local`                    | Boot the local Directus stack, run acceptance tests, tear it down                               |
| `task fmt` / `task fmt:check`           | Format / check formatting                                                                       |
| `task vet`                              | Run `go vet`                                                                                    |
| `task lint`                             | Run `golangci-lint`                                                                             |
| `task tidy`                             | Tidy & verify `go.mod`/`go.sum`                                                                 |
| `task generate` / `task docs`           | Regenerate registry docs via `tfplugindocs`                                                     |
| `task compose:up` / `task compose:down` | Start / stop the local Directus stack                                                           |
| `task ci`                               | Full local gate: fmt check, vet, lint, tests                                                    |

### Acceptance tests

Acceptance tests run against a live Directus. The repo ships a local
[`docker-compose.yml`](docker-compose.yml) (Postgres + Directus, response cache
disabled for deterministic test refreshes):

```sh
task testacc:local
```

Or point the tests at any instance:

```sh
TF_ACC=1 DIRECTUS_URL=... DIRECTUS_TOKEN=... go test ./internal/provider/ -run TestAcc
```

## License

See [LICENSE](LICENSE).
