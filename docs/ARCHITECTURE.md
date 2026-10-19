# Architecture

This driver uses the **Cloudflare Email Service REST API**, instead of SMTP or the Workers binding. This hides the implementation details behind a single package. Switching to another method touches only this repository, adhering to the open/closed principle.

Cloudflare Email Sending is currently in beta.

## Out of Scope

The following features are not supported in this version:
- HTML content
- Attachments
- Retries (retries belong to `alert_manager`)
- Queues
- Templates
