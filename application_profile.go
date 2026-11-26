package custd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	currentAppProfilePath = "/api/v1/app/profile"
	currentAppPhotoPath   = currentAppProfilePath + "/photo"
	currentAppPhotoLimit  = 256 * 1024
)

// CurrentAppProfile is the profile disclosed to the application's signed-in
// member. Fields report whether the member can edit them. ManagedPhoto means
// the uploaded photo must be read with the same app profile grant.
type CurrentAppProfile struct {
	Fields       []CurrentAppProfileField `json:"fields"`
	ManagedPhoto bool                     `json:"managedPhoto,omitempty"`
}

// CurrentAppProfileField includes whether a field has ever been written.
// A cleared field remains initialized so importing old app data cannot restore it.
type CurrentAppProfileField struct {
	ApplicationProfileField
	Initialized bool `json:"initialized"`
}

type CurrentAppProfilePhoto struct {
	ContentType string
	Bytes       []byte
}

func (c *CustdClient) GetCurrentAppProfile(ctx context.Context, token string) (*CurrentAppProfile, error) {
	return c.currentAppProfile(ctx, token, http.MethodGet, currentAppProfilePath, "", nil)
}

func (c *CustdClient) UpdateCurrentAppProfile(ctx context.Context, token string, update ApplicationProfileValuesRequest) (*CurrentAppProfile, error) {
	return c.writeCurrentAppProfile(ctx, token, http.MethodPatch, update)
}

// InitializeCurrentAppProfile fills only previously uninitialized fields.
// The response carries the current values, including any concurrent winner.
func (c *CustdClient) InitializeCurrentAppProfile(ctx context.Context, token string, update ApplicationProfileValuesRequest) (*CurrentAppProfile, error) {
	return c.writeCurrentAppProfile(ctx, token, http.MethodPost, update)
}

func (c *CustdClient) writeCurrentAppProfile(ctx context.Context, token, method string, update ApplicationProfileValuesRequest) (*CurrentAppProfile, error) {
	body, err := json.Marshal(update)
	if err != nil {
		return nil, fmt.Errorf("custd: encode application profile update: %w", err)
	}
	return c.currentAppProfile(ctx, token, method, currentAppProfilePath, "application/json", body)
}

func (c *CustdClient) GetCurrentAppProfilePhoto(ctx context.Context, token string) (*CurrentAppProfilePhoto, error) {
	response, err := c.requestCurrentAppProfile(ctx, token, http.MethodGet, currentAppPhotoPath, "", nil)
	if err != nil {
		return nil, err
	}
	for name, value := range response.Headers {
		if strings.EqualFold(name, "Content-Type") {
			photo := &CurrentAppProfilePhoto{ContentType: value, Bytes: response.Body}
			if err := photo.validate(); err != nil {
				return nil, err
			}
			return photo, nil
		}
	}
	return nil, validationErrorf("custd: current app profile photo content type is required")
}

func (c *CustdClient) UploadCurrentAppProfilePhoto(ctx context.Context, token string, photo CurrentAppProfilePhoto) (*CurrentAppProfile, error) {
	return c.writeCurrentAppProfilePhoto(ctx, token, http.MethodPut, photo)
}

// InitializeCurrentAppProfilePhoto fills only a previously uninitialized photo.
func (c *CustdClient) InitializeCurrentAppProfilePhoto(ctx context.Context, token string, photo CurrentAppProfilePhoto) (*CurrentAppProfile, error) {
	return c.writeCurrentAppProfilePhoto(ctx, token, http.MethodPost, photo)
}

func (c *CustdClient) writeCurrentAppProfilePhoto(ctx context.Context, token, method string, photo CurrentAppProfilePhoto) (*CurrentAppProfile, error) {
	if err := photo.validate(); err != nil {
		return nil, err
	}
	return c.currentAppProfile(ctx, token, method, currentAppPhotoPath, photo.ContentType, photo.Bytes)
}

func (c *CustdClient) RemoveCurrentAppProfilePhoto(ctx context.Context, token string) (*CurrentAppProfile, error) {
	return c.currentAppProfile(ctx, token, http.MethodDelete, currentAppPhotoPath, "", nil)
}

func (p CurrentAppProfilePhoto) validate() error {
	if (p.ContentType != "image/png" && p.ContentType != "image/webp") || len(p.Bytes) == 0 || len(p.Bytes) > currentAppPhotoLimit {
		return validationErrorf("custd: app profile photo must be PNG or WebP and at most 256 KiB")
	}
	return nil
}

func (c *CustdClient) currentAppProfile(ctx context.Context, token, method, path, contentType string, body []byte) (*CurrentAppProfile, error) {
	response, err := c.requestCurrentAppProfile(ctx, token, method, path, contentType, body)
	if err != nil {
		return nil, err
	}
	const contextName = "current app profile"
	fields, err := decodeJSONObject(response.Body, contextName)
	if err != nil {
		return nil, err
	}
	if err := requireJSONObjectList(fields, contextName, "fields", "fieldKey", "present", "required", "editableBy", "editable"); err != nil {
		return nil, err
	}
	var profile CurrentAppProfile
	if err := unmarshalResponse(response.Body, &profile); err != nil {
		return nil, validationWrap("custd: decode current app profile", err)
	}
	return &profile, nil
}

func (c *CustdClient) requestCurrentAppProfile(ctx context.Context, token, method, path, contentType string, body []byte) (*HTTPResponse, error) {
	if token == "" || len(strings.Fields(token)) != 1 || strings.TrimSpace(token) != token {
		return nil, fmt.Errorf("custd: application profile access token is required")
	}
	base := c.config.BaseURL
	origin, err := url.Parse(base)
	if err != nil || origin.Hostname() == "" || origin.User != nil || origin.RawQuery != "" ||
		origin.ForceQuery || origin.Fragment != "" || (origin.Path != "" && origin.Path != "/") || !isSecureOrLocalHTTP(base) {
		return nil, fmt.Errorf("custd: current app profile requires an HTTPS API origin")
	}
	headers := map[string]string{"Authorization": "Bearer " + token, "X-Custd-Sdk": sdkIdentity}
	if contentType != "" {
		headers["Content-Type"] = contentType
	}
	response, err := c.requestApplication(ctx, &HTTPRequest{
		Method: method, URL: strings.TrimRight(base, "/") + path, Headers: headers, Body: body,
	})
	if err != nil {
		return nil, err
	}
	if err := c.checkStatus(response.StatusCode, response.Body); err != nil {
		return nil, err
	}
	return response, nil
}
