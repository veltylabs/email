//go:build !wasm

package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"webtyp.com/msgtype"
	"webtyp.com/notify"
)

type channel struct {
	name      string
	accountID string
	from      string
	fromName  string
	endpoint  string
	token     string
	client    *http.Client
}

func (c *channel) Name() string { return c.name }

type sendRequest struct {
	From    sender    `json:"from"`
	To      recipient `json:"to"`
	Subject string    `json:"subject"`
	Text    string    `json:"text"`
}

type sender struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type recipient struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type sendResponse struct {
	Success bool            `json:"success"`
	Errors  []responseError `json:"errors"`
	Result  responseResult  `json:"result"`
}

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type responseResult struct {
	Delivered        []string `json:"delivered"`
	PermanentBounces []string `json:"permanent_bounces"`
	Queued           []string `json:"queued"`
}

// Subject prefixes by alert level; an Info alert has none.
const (
	prefixError   = "[URGENT] "
	prefixWarning = "[WARNING] "
)

func (c *channel) Send(a notify.Alert, to notify.Recipient) error {
	if to.Email == "" {
		return notify.ErrNoAddress
	}

	subject := a.Subject
	switch a.Level {
	case msgtype.Error:
		subject = prefixError + subject
	case msgtype.Warning:
		subject = prefixWarning + subject
	}

	reqBody := sendRequest{
		From: sender{
			Email: c.from,
			Name:  c.fromName,
		},
		To: recipient{
			Email: to.Email,
			Name:  to.Name,
		},
		Subject: subject,
		Text:    a.Body,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("email: cannot marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/accounts/%s/email/sending/send", c.endpoint, c.accountID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("email: cannot create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("email: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("email: cannot read response: %w", err)
	}

	var res sendResponse
	if err := json.Unmarshal(respBody, &res); err != nil {
		return fmt.Errorf("email: cannot unmarshal response: %w (status %d)", err, resp.StatusCode)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 && res.Success {
		for _, e := range res.Result.Delivered {
			if e == to.Email {
				return nil
			}
		}
		for _, e := range res.Result.Queued {
			if e == to.Email {
				return nil
			}
		}
	}

	for _, e := range res.Result.PermanentBounces {
		if e == to.Email {
			return errRecipientRejected
		}
	}

	var errorMsgs string
	for _, e := range res.Errors {
		errorMsgs += fmt.Sprintf(" code %d: %s;", e.Code, e.Message)
	}

	return fmt.Errorf("email: send failed (status %d):%s", resp.StatusCode, errorMsgs)
}
