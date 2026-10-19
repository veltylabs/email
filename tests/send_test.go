//go:build !wasm

package email_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"webtyp.com/notify"
	"github.com/veltylabs/email"
)

func TestSend(t *testing.T) {
	const fakeToken = "super-secret-cloudflare-token-never-log"

	tests := []struct {
		name          string
		alert         notify.Alert
		to            notify.Recipient
		serverStatus  int
		serverBody    string
		sleepDuration time.Duration
		wantError     string
		checkReq      func(t *testing.T, req *http.Request, body string)
	}{
		{
			name: "success delivered",
			alert: notify.Alert{
				Subject: "Hello",
				Body:    "World",
			},
			to:           notify.Recipient{Email: "user@example.com"},
			serverStatus: http.StatusOK,
			serverBody:   `{"success": true, "errors": [], "result": {"delivered": ["user@example.com"], "permanent_bounces": [], "queued": []}}`,
			checkReq: func(t *testing.T, req *http.Request, body string) {
				if req.URL.Path != "/accounts/acc123/email/sending/send" {
					t.Errorf("wrong path: %s", req.URL.Path)
				}
				if req.Header.Get("Authorization") != "Bearer "+fakeToken {
					t.Errorf("wrong authorization header: %s", req.Header.Get("Authorization"))
				}
				if req.Header.Get("Content-Type") != "application/json" {
					t.Errorf("wrong content type")
				}
				if !strings.Contains(body, `"subject":"Hello"`) {
					t.Errorf("wrong subject in body: %s", body)
				}
				if !strings.Contains(body, `"from":{"email":"from@example.com","name":"Test Sender"}`) {
					t.Errorf("wrong from in body: %s", body)
				}
			},
		},
		{
			name: "success queued",
			alert: notify.Alert{
				Subject: "Hello",
			},
			to:           notify.Recipient{Email: "user@example.com"},
			serverStatus: http.StatusOK,
			serverBody:   `{"success": true, "errors": [], "result": {"delivered": [], "permanent_bounces": [], "queued": ["user@example.com"]}}`,
		},
		{
			name: "permanent_bounces",
			alert: notify.Alert{
				Subject: "Hello",
			},
			to:           notify.Recipient{Email: "user@example.com"},
			serverStatus: http.StatusOK,
			serverBody:   `{"success": true, "errors": [], "result": {"delivered": [], "permanent_bounces": ["user@example.com"], "queued": []}}`,
			wantError:    "email: the recipient address was rejected",
		},
		{
			name: "error response 400",
			alert: notify.Alert{
				Subject: "Hello",
			},
			to:           notify.Recipient{Email: "user@example.com"},
			serverStatus: http.StatusBadRequest,
			serverBody:   `{"success": false, "errors": [{"code": 1000, "message": "bad request"}], "result": {}}`,
			wantError:    "status 400",
		},
		{
			name: "error response 500",
			alert: notify.Alert{
				Subject: "Hello",
			},
			to:           notify.Recipient{Email: "user@example.com"},
			serverStatus: http.StatusInternalServerError,
			serverBody:   `{"success": false, "errors": [], "result": {}}`,
			wantError:    "status 500",
		},
		{
			name: "missing to address",
			alert: notify.Alert{
				Subject: "Hello",
			},
			to:        notify.Recipient{},
			wantError: notify.ErrNoAddress.Error(),
		},
		{
			name: "urgency critical",
			alert: notify.Alert{
				Subject: "System Down",
				Urgency: notify.Critical,
			},
			to:           notify.Recipient{Email: "user@example.com"},
			serverStatus: http.StatusOK,
			serverBody:   `{"success": true, "result": {"delivered": ["user@example.com"]}}`,
			checkReq: func(t *testing.T, req *http.Request, body string) {
				if !strings.Contains(body, `"subject":"[URGENT] System Down"`) {
					t.Errorf("wrong subject in body: %s", body)
				}
			},
		},
		{
			name: "urgency warning",
			alert: notify.Alert{
				Subject: "Disk low",
				Urgency: notify.Warning,
			},
			to:           notify.Recipient{Email: "user@example.com"},
			serverStatus: http.StatusOK,
			serverBody:   `{"success": true, "result": {"delivered": ["user@example.com"]}}`,
			checkReq: func(t *testing.T, req *http.Request, body string) {
				if !strings.Contains(body, `"subject":"[WARNING] Disk low"`) {
					t.Errorf("wrong subject in body: %s", body)
				}
			},
		},
		{
			name:          "timeout",
			alert:         notify.Alert{Subject: "Hello"},
			to:            notify.Recipient{Email: "user@example.com"},
			sleepDuration: 20 * time.Millisecond,
			wantError:     "Client.Timeout exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverCalled := false
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serverCalled = true
				if tt.sleepDuration > 0 {
					time.Sleep(tt.sleepDuration)
				}
				bodyBytes, _ := io.ReadAll(r.Body)
				if tt.checkReq != nil {
					tt.checkReq(t, r, string(bodyBytes))
				}
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverBody))
			}))
			defer srv.Close()

			d := &email.Driver{
				Client: &http.Client{Timeout: 10 * time.Millisecond},
			}

			config := notify.Config{
				Settings: []notify.Setting{
					{Key: email.SettingAccountID, Value: "acc123"},
					{Key: email.SettingFrom, Value: "from@example.com"},
					{Key: email.SettingFromName, Value: "Test Sender"},
					{Key: email.SettingEndpoint, Value: srv.URL},
				},
				Secrets: []notify.Setting{
					{Key: email.SecretAPIToken, Value: "the_token"},
				},
			}
			resolver := fakeResolver{"the_token": fakeToken}
			ch, err := d.New(config, resolver)
			if err != nil {
				t.Fatalf("New failed: %v", err)
			}

			err = ch.Send(tt.alert, tt.to)

			if tt.wantError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantError)
				}
				if !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantError)
				}
				if strings.Contains(err.Error(), fakeToken) {
					t.Errorf("error message contains the secret value!")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if tt.name == "missing to address" && serverCalled {
				t.Errorf("server should not have been called")
			}
		})
	}
}
