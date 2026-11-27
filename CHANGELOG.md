## 0.1.0 (Unreleased)

BREAKING CHANGES:
* `vy_app_client.client_secret` is now marked as sensitive. Any `output` that refers to it, also in child modules,
  must set `sensitive = true`, otherwise `terraform plan` fails with "Output refers to sensitive values".
  Release this as a new major version (or a new minor while the provider is still 0.x) and announce it to the teams first.

FEATURES:
* Add resource `vy-resource-server`
* Add resource `vy-app-client`
* Add resource `vy-resource-server-scope`
