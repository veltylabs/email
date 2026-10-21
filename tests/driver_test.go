//go:build !wasm

package email_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/veltylabs/email"
	"webtyp.com/notify"
)

type fakeResolver map[string]string

func (f fakeResolver) Secret(key string) (string, error) {
	if key == "error" {
		return "", errors.New("resolver failure")
	}
	return f[key], nil
}

func TestDriver_Kind(t *testing.T) {
	d := &email.Driver{}
	if d.Kind() != email.Kind {
		t.Errorf("got %q, want %q", d.Kind(), email.Kind)
	}
}

func TestDriver_New(t *testing.T) {
	d := &email.Driver{}
	resolver := fakeResolver{"the_token": "secret123", "empty_token": ""}

	tests := []struct {
		name      string
		config    notify.Config
		wantError string
	}{
		{
			name: "missing account_id",
			config: notify.Config{
				Settings: []notify.Setting{{Key: email.SettingFrom, Value: "test@example.com"}},
				Secrets:  []notify.Setting{{Key: email.SecretAPIToken, Value: "the_token"}},
			},
			wantError: "missing account_id",
		},
		{
			name: "missing from",
			config: notify.Config{
				Settings: []notify.Setting{{Key: email.SettingAccountID, Value: "acc123"}},
				Secrets:  []notify.Setting{{Key: email.SecretAPIToken, Value: "the_token"}},
			},
			wantError: "missing from",
		},
		{
			name: "invalid from - no @",
			config: notify.Config{
				Settings: []notify.Setting{{Key: email.SettingAccountID, Value: "acc123"}, {Key: email.SettingFrom, Value: "testexample.com"}},
				Secrets:  []notify.Setting{{Key: email.SecretAPIToken, Value: "the_token"}},
			},
			wantError: "invalid from",
		},
		{
			name: "invalid from - prefix @",
			config: notify.Config{
				Settings: []notify.Setting{{Key: email.SettingAccountID, Value: "acc123"}, {Key: email.SettingFrom, Value: "@example.com"}},
				Secrets:  []notify.Setting{{Key: email.SecretAPIToken, Value: "the_token"}},
			},
			wantError: "invalid from",
		},
		{
			name: "invalid from - suffix @",
			config: notify.Config{
				Settings: []notify.Setting{{Key: email.SettingAccountID, Value: "acc123"}, {Key: email.SettingFrom, Value: "test@"}},
				Secrets:  []notify.Setting{{Key: email.SecretAPIToken, Value: "the_token"}},
			},
			wantError: "invalid from",
		},
		{
			name: "invalid from - spaces",
			config: notify.Config{
				Settings: []notify.Setting{{Key: email.SettingAccountID, Value: "acc123"}, {Key: email.SettingFrom, Value: "test @example.com"}},
				Secrets:  []notify.Setting{{Key: email.SecretAPIToken, Value: "the_token"}},
			},
			wantError: "invalid from",
		},
		{
			name: "missing secret key",
			config: notify.Config{
				Settings: []notify.Setting{{Key: email.SettingAccountID, Value: "acc123"}, {Key: email.SettingFrom, Value: "test@example.com"}},
			},
			wantError: "missing secret key api_token",
		},
		{
			name: "resolver error",
			config: notify.Config{
				Settings: []notify.Setting{{Key: email.SettingAccountID, Value: "acc123"}, {Key: email.SettingFrom, Value: "test@example.com"}},
				Secrets:  []notify.Setting{{Key: email.SecretAPIToken, Value: "error"}},
			},
			wantError: "resolver failure",
		},
		{
			name: "empty token resolved",
			config: notify.Config{
				Settings: []notify.Setting{{Key: email.SettingAccountID, Value: "acc123"}, {Key: email.SettingFrom, Value: "test@example.com"}},
				Secrets:  []notify.Setting{{Key: email.SecretAPIToken, Value: "empty_token"}},
			},
			wantError: "resolved secret api_token is empty",
		},
		{
			name: "success",
			config: notify.Config{
				Name:     "my_email",
				Settings: []notify.Setting{{Key: email.SettingAccountID, Value: "acc123"}, {Key: email.SettingFrom, Value: "test@example.com"}},
				Secrets:  []notify.Setting{{Key: email.SecretAPIToken, Value: "the_token"}},
			},
			wantError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, err := d.New(tt.config, resolver)
			if tt.wantError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantError)
				}
				if !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantError)
				}
				if ch != nil {
					t.Errorf("expected nil channel on error")
				}
				if strings.Contains(err.Error(), "secret123") {
					t.Errorf("error message contains the secret value")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if ch == nil {
					t.Fatalf("expected channel, got nil")
				}
				if ch.Name() != tt.config.Name {
					t.Errorf("got name %q, want %q", ch.Name(), tt.config.Name)
				}
			}
		})
	}
}

func TestDriver_Register(t *testing.T) {
	r := &notify.Registry{}
	if err := r.RegisterDriver(&email.Driver{}); err != nil {
		t.Fatalf("F1 failed: RegisterDriver returned err: %v", err)
	}

	label, fields, ok := r.DriverOf(email.Kind)
	if !ok {
		t.Fatalf("F2 failed: DriverOf returned ok=false")
	}
	if label != "Email" {
		t.Errorf("F2 failed: label is %q, want \"Email\"", label)
	}
	if len(fields) != 4 {
		t.Fatalf("F2 failed: expected 4 fields, got %d", len(fields))
	}
	if fields[0].Key != email.SettingAccountID || fields[1].Key != email.SettingFrom || fields[2].Key != email.SettingFromName || fields[3].Key != email.SecretAPIToken {
		t.Errorf("F2 failed: fields order incorrect")
	}

	config := notify.Config{
		Name:     "test_email",
		Kind:     email.Kind,
		Enabled:  true,
		Settings: []notify.Setting{{Key: email.SettingAccountID, Value: "acc"}, {Key: email.SettingFrom, Value: "test@example.com"}},
		Secrets:  []notify.Setting{{Key: email.SecretAPIToken, Value: "the_token"}},
	}
	resolver := fakeResolver{"the_token": "secret"}

	err := r.Load([]notify.Config{config}, resolver)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !r.Has("test_email") {
		t.Fatal("Channel not found in registry")
	}
}

func TestDriver_Fields(t *testing.T) {
	fields := (&email.Driver{}).Fields()

	requiredMap := map[string]bool{}
	secretMap := map[string]bool{}
	hasEndpoint := false

	for _, f := range fields {
		if f.Required {
			requiredMap[f.Key] = true
		}
		if f.Secret {
			secretMap[f.Key] = true
		}
		if f.Key == email.SettingEndpoint {
			hasEndpoint = true
		}
	}

	if !requiredMap[email.SettingAccountID] || !requiredMap[email.SettingFrom] || !requiredMap[email.SecretAPIToken] {
		t.Errorf("F3 failed: account_id, from, and api_token should be Required")
	}

	if len(secretMap) != 1 || !secretMap[email.SecretAPIToken] {
		t.Errorf("F3 failed: api_token should be the only Secret")
	}

	if hasEndpoint {
		t.Errorf("F3 failed: field with key endpoint should not exist")
	}
}

func TestDriver_New_NilResolver(t *testing.T) {
	d := &email.Driver{}
	cfg := notify.Config{
		Settings: []notify.Setting{{Key: email.SettingAccountID, Value: "acc123"}, {Key: email.SettingFrom, Value: "test@example.com"}},
		Secrets:  []notify.Setting{{Key: email.SecretAPIToken, Value: "the_token"}},
	}
	if _, err := d.New(cfg, nil); err == nil {
		t.Fatal("a nil SecretResolver must be an error, not a panic")
	}
}
