# Architecture

This driver uses the **Cloudflare Email Service REST API**, instead of SMTP or the Workers binding. This hides the implementation details behind a single package. Switching to another method touches only this repository, adhering to the open/closed principle.

Cloudflare Email Sending is currently in beta.

## Configuration Fields

| Key | Label | Secret | Required |
|---|---|---|---|
| `account_id` | Cloudflare account ID | No | Yes |
| `from` | Sender address | No | Yes |
| `from_name` | Sender name | No | No |
| `api_token` | API token secret name | Yes | Yes |

The `endpoint` setting is purposefully not declared as a field because it exists only so tests can point the driver at `httptest`, and a person must never see it.

## Out of Scope

The following features are not supported in this version:
- HTML content
- Attachments
- Retries (retries belong to `alert_manager`)
- Queues
- Templates
