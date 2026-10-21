# email
<img src="docs/img/badges.svg">

The email channel for webtyp/notify. It sends alerts through the Cloudflare Email Service.

## Configuration

It is configured through the application's database (`notify.Config`).

| Key | Label | Secret | Required |
|---|---|---|---|
| `account_id` | Cloudflare account ID | No | Yes |
| `from` | Sender address | No | Yes |
| `from_name` | Sender name | No | No |
| `api_token` | API token secret name | Yes | Yes |

The `endpoint` setting is purposefully not declared as a field because it exists only so tests can point the driver at `httptest`, and a person must never see it.

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
