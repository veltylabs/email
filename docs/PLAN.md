---
PLAN: "feat!: the email driver declares its label and fields (notify.Driver.Label/Fields)"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 16563578987340549225
---

> Read AGENTS.md at the repo root before any change.

# email: declare the data the driver needs

## Why

`webtyp.com/notify` `v0.2.3` changed the `Driver` contract: a driver now declares a label and
the fields it reads, so one generic «Sending channels» screen can ask for them. This driver reads
`account_id`, `from`, `from_name` and the secret `api_token` (`email.go`); it must declare them.
`endpoint` stays undeclared on purpose: it exists only so tests can point the driver at
`httptest`, and a person must never see it.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
```

## Environment

```bash
export GOPRIVATE=webtyp.com/*,github.com/veltylabs/*
export GOFLAGS=-mod=mod
go get webtyp.com/notify@v0.2.3 webtyp.com/lang@v0.1.7
```

A PR that does not compile, or whose tests could not run, is a failed delivery. Never pause to ask
permission or report progress: when the Delivery check passes, commit, push and open the PR.

## APIs used

| Symbol | Exact shape | Tag |
|---|---|---|
| `notify.Driver` | `interface{ Kind() string; Label() lang.Text; Fields() []Field; New(c Config, secrets SecretResolver) (Channel, error) }` | notify `v0.2.3` |
| `notify.Field` | `struct{ Key string; Label lang.Text; Help lang.Text; Secret bool; Required bool }` | notify `v0.2.3` |
| `notify.Registry.RegisterDriver` | `func (r *Registry) RegisterDriver(d Driver) error` | notify `v0.2.3` |
| `notify.Registry.DriverOf` | `func (r *Registry) DriverOf(kind string) (label lang.Text, fields []Field, ok bool)` | notify `v0.2.3` |

If a signature differs, stop and name it in the PR; never adapt it by guessing.

## Stage 1 — `email.go`

```go
func (d *Driver) Label() lang.Text { return "Email" }

func (d *Driver) Fields() []notify.Field {
	return []notify.Field{
		{Key: SettingAccountID, Label: "Cloudflare account ID", Help: "Cloudflare dashboard → Account home → Account ID", Required: true},
		{Key: SettingFrom, Label: "Sender address", Help: "An address on a domain onboarded in Cloudflare Email Sending", Required: true},
		{Key: SettingFromName, Label: "Sender name"},
		{Key: SecretAPIToken, Label: "API token secret name", Help: "Name of the secret that holds a Cloudflare API token with Email Sending permission", Secret: true, Required: true},
	}
}
```

`New` is unchanged (it still validates; `Registry.Load` now also refuses a config missing a
required field before calling `New`).

## Stage 2 — translations

`langc sync .` creates `lang.json` (positional, `es` column). Fill:

```json
"API token secret name": ["Nombre del secreto del token de API"],
"Cloudflare account ID": ["ID de la cuenta de Cloudflare"],
"Email": ["Correo electrónico"],
```

…and every other key `langc` adds (the `Help` texts and the remaining labels), all with `es`.

## Stage 3 — tests (`tests/driver_test.go`)

| # | Setup | Call | Expected |
|---|---|---|---|
| F1 | `r := &notify.Registry{}` | `r.RegisterDriver(&email.Driver{})` | `nil` (the fields pass the registry's checks) |
| F2 | F1 | `r.DriverOf(email.Kind)` | `ok`, label `"Email"`, 4 fields in the order above |
| F3 | — | `(&email.Driver{}).Fields()` | `account_id`, `from`, `api_token` are `Required`; `api_token` is the only `Secret`; no field has key `endpoint` |

Existing tests stay green.

## Stage 4 — docs

`README.md` and `docs/ARCHITECTURE.md`: the table of fields (key, label, secret, required) and the
sentence on why `endpoint` is not declared.

## Delivery check

```bash
go build ./...
gotest
git diff --stat main   # email.go, lang.json, tests/driver_test.go, go.mod, go.sum, README.md, docs/ARCHITECTURE.md
```
