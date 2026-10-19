---
PLAN: "feat: email v0.1.0 — the Cloudflare Email Service channel for webtyp/notify (driver kind \\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\"email\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\")"
TAG: v0.1.0
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 3840922618613193728
PR: https://github.com/veltylabs/email/pull/1
REVIEWED_COMMIT: eb83c96b1ff0bb867db3b0f4e654d8f8bb705f0f
---

> `webtyp.com/notify` v0.1.0 is published; this plan depends on it.

> Read AGENTS.md at the repo root before any change.
> This is a **new library**. `gonew` created only `go.mod`, `LICENSE`, `README.md` and a stub `email.go`.

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
```

## What this library is

The first concrete channel of `webtyp.com/notify`: it sends an `Alert` as an email through **Cloudflare Email Service**
(Email Sending, public beta). It implements `notify.Driver` (kind `"email"`) so that an instance of this channel is a
**row in the application's database** (`notify.Config`), created or changed without a rebuild. It adds no concept of its
own. Prices and limits for the owner are in `docs/PRICING.md` (already written; do not edit it).

Dependency: `webtyp.com/notify` at `v0.1.0`. Server-side only: every non-test file carries `//go:build !wasm`.
Because this is integration code for an outside service, the standard library's `net/http` and `encoding/json` are
allowed here (follow how `veltylabs/transbank` does its HTTP calls).

## Design gate

1. **Prior art.** (a) Django's `EmailBackend` and Laravel mail transports: one interface, one class per service,
   chosen by configuration. (b) `log/slog.Handler`: one small interface injected at the edge. (c) Cloudflare's own
   Workers binding, REST API and SMTP: three ways to reach the same service — this package uses the **REST API** and
   hides that choice, so switching to SMTP or another provider touches only this repo.
2. **Novice-name test.** `email.Driver`, `email.Kind`, setting keys `account_id`, `from`, `from_name`, `endpoint`,
   secret key `api_token`. Read aloud: "an email channel that sends from this address through this account".
3. **Complexity ledger.** Concepts +1 (the driver); lines at the call site −25 (HTTP, auth, error mapping written once);
   ways to send an alert by email: +0 / −0 (the first one).
4. **Where it belongs.** Its own repo: one concern (this service). A second email provider is another repo implementing
   the same `notify.Driver`; it touches zero lines here (open/closed).
5. **What it deletes.** Nothing (new).

## Public API (package `email`)

```go
// Kind is the notify.Driver kind this package builds.
const Kind = "email"

// Setting and secret keys read from notify.Config.
const (
	SettingAccountID = "account_id" // required: the Cloudflare account id
	SettingFrom      = "from"       // required: the sender address, on a domain onboarded in Email Sending
	SettingFromName  = "from_name"  // optional: the sender's display name
	SettingEndpoint  = "endpoint"   // optional: base URL of the API; default DefaultEndpoint. Tests point it at httptest.
	SecretAPIToken   = "api_token"  // required: the name under which notify.Config.Secrets holds the token's NAME in the secret store
)

const DefaultEndpoint = "https://api.cloudflare.com/client/v4"

// Driver implements notify.Driver. The zero value is ready to use (it creates an http.Client with a 15-second timeout).
type Driver struct {
	Client *http.Client // optional
}

func (d *Driver) Kind() string
func (d *Driver) New(c notify.Config, secrets notify.SecretResolver) (notify.Channel, error)
```

`New` returns `notify.ErrConfig` wrapped in a message that names the **key** (never a value) when `account_id`, `from` or
`Secrets[api_token]` is missing or blank, when `from` is not a plausible address (one `@`, non-empty both sides, no
spaces), or when the secret cannot be resolved (the resolver's error text is allowed; the secret value never is). The
returned channel's `Name()` equals `c.Name`.

### What `Send(a notify.Alert, to notify.Recipient)` does

1. `to.Email` empty → `notify.ErrNoAddress`.
2. `POST {endpoint}/accounts/{account_id}/email/sending/send` with headers `Authorization: Bearer <token>` and
   `Content-Type: application/json`, and a JSON body:
   `{"from": {"email": <from>, "name": <from_name, omitted if empty>}, "to": {"email": <to.Email>, "name": <to.Name, omitted if empty>}, "subject": <subject>, "text": <a.Body>}`.
   The subject is `a.Subject`, prefixed with `"[URGENT] "` when `a.Urgency == notify.Critical` and `"[WARNING] "` when
   `notify.Warning` (two named constants). Text only; no HTML in this version.
3. Read the response: `{"success": bool, "errors": [{"code": n, "message": s}], "result": {"delivered": [..], "permanent_bounces": [..], "queued": [..]}}`.
   - HTTP 2xx, `success == true` and the recipient in `delivered` or `queued` → return nil.
   - The recipient in `permanent_bounces` → return an `email.Error` with the constant text `"email: the recipient address was rejected"`.
   - Any other outcome → return an error whose text contains the HTTP status and every `errors[].message`, never the token or the request headers.
4. A network error or a timeout is returned as an error (never swallowed, never retried here: retries belong to `alert_manager`).

`email.Error` is `type Error string` with the constants used above, compared by value. No `map`, no `reflect`, no
`log` of request or response bodies, no clock beyond the client timeout.

## Stages

### Stage 1 — the driver (`email.go` replacing the stub, `send.go`)

As specified. `go.mod`: `go 1.26.8`, requires `webtyp.com/notify` at `v0.1.0`; no local `replace`.

### Stage 2 — tests (`tests/`, external package `email_test`, standard library incl. `net/http/httptest`)

Through the exported API, with a `httptest.Server` standing for Cloudflare and `notifytest`-style fakes for the resolver:
- `New`: each missing or blank key (`account_id`, `from`, `api_token`), a bad `from`, a resolver error, and success; `Name()` equals the config name.
- A driver built from a `notify.Config` works through `notify.Registry.Load` (registered with `RegisterDriver`): the channel appears and `Deliver` reaches the test server.
- `Send`: success (`delivered`), success (`queued`), `permanent_bounces`, a 400 with `errors`, a 500, a timeout (a server that sleeps past a short client timeout injected with `Driver.Client`), no `to.Email` (`ErrNoAddress`, and the server is never called).
- The request: path with the account id, the `Authorization: Bearer` header, the JSON body fields, the subject prefixes for `Warning` and `Critical`, `name` omitted when empty.
- **The token never appears** in any returned error text (assert with a recognisable fake token).

### Stage 3 — docs

`README.md`: what it is, how an instance is configured as a `notify.Config` row (the settings, and that the secret is
referenced **by name**), a 15-line example registering the driver and loading one config, and a link to
`docs/PRICING.md`. `docs/ARCHITECTURE.md`: why REST and not SMTP or the Workers binding, the beta status of Email Sending,
what is out of scope (HTML, attachments, retries, queues, templates). Permanent docs never link `docs/PLAN.md`.

## Delivery check

```
gotest                                                       # vet, race and tests green
grep -rn "map\[\|reflect\|\"errors\"\|\"strings\"\|\"strconv\"" --include=*.go . | grep -v _test   # → empty
grep -n '^replace' go.mod                                    # → empty
git diff --stat main                                         # only the files of the stages
```

## Stages table

| # | Stage | Files |
|---|---|---|
| 1 | the driver | `email.go`, `send.go`, `go.mod` |
| 2 | tests | `tests/driver_test.go`, `tests/send_test.go` |
| 3 | docs | `README.md`, `docs/ARCHITECTURE.md` |

## Executor notes
- `webtyp.com/notify` at `v0.1.0` defines `notify.Config.Settings` and `notify.Config.Secrets` as `[]notify.Setting`, not maps. Iterated over these to get values using a helper function.
- `SecretResolver` interface in `v0.1.0` has a `Secret(name string) (string, error)` method, not `Resolve()`.
- Implemented `contains`, `hasPrefix`, and `hasSuffix` manually in `email.go` to conform with rule prohibiting `strings` standard package.
- Used `http.NewRequest` instead of `http.NewRequestWithContext` since `Channel.Send` method in `notify` interface doesn't accept a `context.Context` parameter. Checked timeout in tests using the injected HTTP Client's internal timeout instead of `context.WithDeadline`.
