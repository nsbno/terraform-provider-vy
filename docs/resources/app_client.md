---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "vy_app_client Resource - terraform-provider-vy"
subcategory: ""
description: |-
  An app client, used to access resource servers.
---

# vy_app_client (Resource)

An app client, used to access resource servers.

## Example Usage

```terraform
resource "vy_app_client" "test" {
  name = "app_client_basic.acceptancetest.io"
  type = "backend"
  scopes = [
    "my.cool.service.vydev.io/read"
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **name** (String) The name of this app client
- **type** (String) The use-case for this app client. Used to automatically add OAuth options

### Optional

- **scopes** (List of String) Scopes that this client has access to

### Read-Only

- **id** (String) The ID of this resource.

