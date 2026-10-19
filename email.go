//go:build !wasm

package email

import (
	"fmt"
	"net/http"
	"time"

	"webtyp.com/notify"
)

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

// Error represents an error returned by the email driver.
type Error string

func (e Error) Error() string { return string(e) }

const (
	errRecipientRejected = Error("email: the recipient address was rejected")
)

// Driver implements notify.Driver. The zero value is ready to use (it creates an http.Client with a 15-second timeout).
type Driver struct {
	Client *http.Client // optional
}

func (d *Driver) Kind() string { return Kind }

func getSetting(settings []notify.Setting, key string) string {
	for _, s := range settings {
		if s.Key == key {
			return s.Value
		}
	}
	return ""
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func (d *Driver) New(c notify.Config, secrets notify.SecretResolver) (notify.Channel, error) {
	accountID := getSetting(c.Settings, SettingAccountID)
	if accountID == "" {
		return nil, fmt.Errorf("%w: missing %s", notify.ErrConfig, SettingAccountID)
	}

	from := getSetting(c.Settings, SettingFrom)
	if from == "" {
		return nil, fmt.Errorf("%w: missing %s", notify.ErrConfig, SettingFrom)
	}
	if !contains(from, "@") || hasPrefix(from, "@") || hasSuffix(from, "@") || contains(from, " ") {
		return nil, fmt.Errorf("%w: invalid %s", notify.ErrConfig, SettingFrom)
	}

	tokenName := getSetting(c.Secrets, SecretAPIToken)
	if tokenName == "" {
		return nil, fmt.Errorf("%w: missing secret key %s", notify.ErrConfig, SecretAPIToken)
	}

	token, err := secrets.Secret(tokenName)
	if err != nil {
		return nil, fmt.Errorf("%w: cannot resolve secret %s: %v", notify.ErrConfig, SecretAPIToken, err)
	}
	if token == "" {
		return nil, fmt.Errorf("%w: resolved secret %s is empty", notify.ErrConfig, SecretAPIToken)
	}

	endpoint := getSetting(c.Settings, SettingEndpoint)
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}

	client := d.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	return &channel{
		name:      c.Name,
		accountID: accountID,
		from:      from,
		fromName:  getSetting(c.Settings, SettingFromName),
		endpoint:  endpoint,
		token:     token,
		client:    client,
	}, nil
}
