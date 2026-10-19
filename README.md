# email
<img src="docs/img/badges.svg">

The email channel for webtyp/notify. It sends alerts through the Cloudflare Email Service.

## Configuration

It is configured through the application's database (`notify.Config`). Settings are:

- `account_id` (required): the Cloudflare account id.
- `from` (required): the sender address, on a domain onboarded in Email Sending.
- `from_name` (optional): the sender's display name.
- `endpoint` (optional): base URL of the API. Defaults to the Cloudflare v4 API.

Secrets are referenced by name, not by value:

- `api_token` (required): the name of the secret in the secret store.

See [docs/PRICING.md](docs/PRICING.md) for costs.

## Example

```go
package main

import (
	"log"

	"github.com/veltylabs/email"
	"webtyp.com/notify"
)

func main() {
	var reg notify.Registry
	if err := reg.RegisterDriver(&email.Driver{}); err != nil {
		log.Fatal(err)
	}

	config := notify.Config{
		Name:    "my_email",
		Kind:    email.Kind,
		Enabled: true,
		Settings: []notify.Setting{
			{Key: email.SettingAccountID, Value: "account_id_here"},
			{Key: email.SettingFrom, Value: "alerts@example.com"},
		},
		Secrets: []notify.Setting{
			{Key: email.SecretAPIToken, Value: "cf_email_token"},
		},
	}

	// SecretResolver implementation omitted for brevity.
	if err := reg.Load([]notify.Config{config}, mySecretResolver); err != nil {
		log.Fatal(err)
	}
}
```
