package custd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const appSettingsEntryPath = "/settings/entry"

// AppAccountEntryRequest binds a ticket to a registered application origin.
// Use the issuer and application origin from trusted deployment configuration.
type AppAccountEntryRequest struct {
	Issuer            string
	ApplicationOrigin string
}

// AppAccountEntry is for a same-tab form POST. The ticket is a short-lived,
// one-use credential and must stay out of URLs and logs.
type AppAccountEntry struct {
	ActionURL string `json:"actionUrl"`
	Ticket    string `json:"ticket"`
}

// CreateAppAccountEntry obtains a hosted settings ticket for a remembered app session.
func (c *CustdClient) CreateAppAccountEntry(ctx context.Context, token string, request AppAccountEntryRequest) (*AppAccountEntry, error) {
	if token == "" || strings.ContainsAny(token, " \t\r\n") {
		return nil, fmt.Errorf("custd: application access token is required")
	}
	issuer, err := accountEntryOrigin(strings.TrimSuffix(request.Issuer, "/"))
	if err != nil {
		return nil, err
	}
	applicationOrigin, err := accountEntryOrigin(request.ApplicationOrigin)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(struct {
		ApplicationOrigin string `json:"applicationOrigin"`
	}{ApplicationOrigin: applicationOrigin})
	if err != nil {
		return nil, fmt.Errorf("custd: encode app account entry request: %w", err)
	}
	action := issuer + appSettingsEntryPath
	response, err := c.requestApplication(ctx, &HTTPRequest{
		Method: http.MethodPost, URL: action + "/ticket",
		Headers: map[string]string{
			"Authorization": "Bearer " + token, "Content-Type": "application/json", "X-Custd-Sdk": sdkIdentity,
		},
		Body: body,
	})
	if err != nil {
		return nil, err
	}
	// Keep a malformed issuer error body from exposing ticket material.
	if err := c.checkStatus(response.StatusCode, nil); err != nil {
		return nil, err
	}
	var entry AppAccountEntry
	if err := unmarshalResponse(response.Body, &entry); err != nil ||
		entry.ActionURL != action || !validAppAccountEntryTicket(entry.Ticket) {
		return nil, validationErrorf("custd: invalid app account entry response")
	}
	return &entry, nil
}

func accountEntryOrigin(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil ||
		"https://"+parsed.Host != raw || parsed.Hostname() != strings.ToLower(parsed.Hostname()) || parsed.Port() == "443" {
		return "", fmt.Errorf("custd: app account entry requires a serialized HTTPS origin")
	}
	return raw, nil
}

func validAppAccountEntryTicket(ticket string) bool {
	if len(ticket) != base64.RawURLEncoding.EncodedLen(32) {
		return false
	}
	data, err := base64.RawURLEncoding.Strict().DecodeString(ticket)
	return err == nil && len(data) == 32
}
