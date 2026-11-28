package custd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const appMessagePath = "/api/v1/app/messages"
const maxAppMessageMediaBytes = 10 << 20

// AppMessageClient operates on the signed-in tenant user's private archive.
type AppMessageClient struct {
	client *CustdClient
}

type AppMessageSource string

const (
	AppMessageSourceSMS      AppMessageSource = "sms"
	AppMessageSourceIMessage AppMessageSource = "imessage"
	AppMessageSourceWhatsApp AppMessageSource = "whatsapp"
)

type AppMessageSession struct {
	CompanySlug              string `json:"companySlug"`
	PrincipalID              string `json:"principalId"`
	ImportAllowed            bool   `json:"importAllowed"`
	WhatsAppSourceGeneration int64  `json:"whatsappSourceGeneration"`
}

type AppMessageInput struct {
	ExternalID   string   `json:"externalId"`
	ThreadID     *string  `json:"threadId"`
	Sender       *string  `json:"sender"`
	Participants []string `json:"participants"`
	Timestamp    string   `json:"timestamp"`
	Direction    string   `json:"direction"`
	Text         string   `json:"text"`
}

type AppMessageBatchRequest struct {
	Source           AppMessageSource  `json:"source"`
	Messages         []AppMessageInput `json:"messages"`
	SourceGeneration *int64            `json:"sourceGeneration,omitempty"`
}

type AppMessageImportResult struct {
	ExternalID  string `json:"externalId"`
	Status      string `json:"status"`
	MessageUUID string `json:"messageUuid,omitempty"`
	ErrorCode   string `json:"errorCode,omitempty"`
}

type AppMessageBatchResponse struct {
	Results []AppMessageImportResult `json:"results"`
}

type AppMessageSearchRequest struct {
	Query    string           `json:"query,omitempty"`
	Sender   string           `json:"sender,omitempty"`
	Source   AppMessageSource `json:"source,omitempty"`
	ThreadID string           `json:"threadId,omitempty"`
	From     *time.Time       `json:"from,omitempty"`
	To       *time.Time       `json:"to,omitempty"`
	Limit    int              `json:"limit,omitempty"`
	Cursor   string           `json:"cursor,omitempty"`
}

type AppMessageMediaMetadata struct {
	MediaUUID   string `json:"mediaUuid"`
	ContentType string `json:"contentType"`
	ByteCount   int64  `json:"byteCount"`
	OCRStatus   string `json:"ocrStatus"`
	OCRText     string `json:"ocrText,omitempty"`
}

type AppMessage struct {
	MessageUUID  string                    `json:"messageUuid"`
	Source       AppMessageSource          `json:"source"`
	ExternalID   string                    `json:"externalId"`
	ThreadID     string                    `json:"threadId"`
	Sender       string                    `json:"sender"`
	Participants []string                  `json:"participants"`
	Timestamp    time.Time                 `json:"timestamp"`
	Direction    string                    `json:"direction"`
	Text         string                    `json:"text"`
	Media        []AppMessageMediaMetadata `json:"media,omitempty"`
}

type AppMessageSearchResponse struct {
	Messages   []AppMessage `json:"messages"`
	NextCursor string       `json:"nextCursor,omitempty"`
}

type AppMessageReadResponse struct {
	Message AppMessage `json:"message"`
}

type AppMessageDeleteSourceResponse struct {
	Source           AppMessageSource `json:"source"`
	DeletedCount     int64            `json:"deletedCount"`
	SourceGeneration int64            `json:"sourceGeneration,omitempty"`
}

type AppMessageMediaUpload struct {
	MessageUUID      string
	Image            []byte
	SourceGeneration *int64
}

type AppMessageMediaRef struct {
	MessageUUID string
	MediaUUID   string
}

type AppMessageMedia struct {
	ContentType string
	Image       []byte
}

func (c *AppMessageClient) Session(ctx context.Context) (*AppMessageSession, error) {
	var out AppMessageSession
	err := c.jsonRequest(ctx, http.MethodGet, "/session", nil, &out, "companySlug", "principalId", "importAllowed")
	return &out, err
}

// Import sends one bounded batch and returns the outcome of each submitted message.
func (c *AppMessageClient) Import(ctx context.Context, req *AppMessageBatchRequest) (*AppMessageBatchResponse, error) {
	var out AppMessageBatchResponse
	err := c.jsonRequest(ctx, http.MethodPost, "/batch", req, &out, "results")
	return &out, err
}

func (c *AppMessageClient) Search(ctx context.Context, req *AppMessageSearchRequest) (*AppMessageSearchResponse, error) {
	var out AppMessageSearchResponse
	err := c.jsonRequest(ctx, http.MethodPost, "/search", req, &out, "messages")
	return &out, err
}

func (c *AppMessageClient) Get(ctx context.Context, messageUUID string) (*AppMessageReadResponse, error) {
	var out AppMessageReadResponse
	err := c.jsonRequest(ctx, http.MethodGet, "/"+url.PathEscape(messageUUID), nil, &out, "message")
	return &out, err
}

// DeleteSource deletes all archived data for one source owned by the signed-in user.
func (c *AppMessageClient) DeleteSource(ctx context.Context, source AppMessageSource) (*AppMessageDeleteSourceResponse, error) {
	var out AppMessageDeleteSourceResponse
	err := c.jsonRequest(ctx, http.MethodDelete, "/sources/"+url.PathEscape(string(source)), nil, &out, "source", "deletedCount")
	return &out, err
}

func (c *AppMessageClient) UploadMedia(ctx context.Context, upload *AppMessageMediaUpload) (*AppMessageMediaMetadata, error) {
	if upload == nil || len(upload.Image) == 0 || len(upload.Image) > maxAppMessageMediaBytes {
		return nil, fmt.Errorf("custd: archive image must contain 1 to %d bytes", maxAppMessageMediaBytes)
	}
	request := c.request(http.MethodPut, "/"+url.PathEscape(upload.MessageUUID)+"/media", upload.Image)
	request.Headers["Content-Type"] = "application/octet-stream"
	if upload.SourceGeneration != nil {
		if *upload.SourceGeneration < 1 {
			return nil, validationErrorf("custd: archive source generation must be positive")
		}
		request.Headers[appMessageSourceGenerationHeader] = strconv.FormatInt(*upload.SourceGeneration, 10)
	}
	_, body, err := c.client.exchangeChecked(ctx, "archive image upload", request)
	if err != nil {
		return nil, err
	}
	var out AppMessageMediaMetadata
	err = decodeAppMessageResponse(body, &out, "mediaUuid", "contentType", "byteCount", "ocrStatus")
	return &out, err
}

// DownloadMedia reads one original image, bounded by the archive's maximum size.
func (c *AppMessageClient) DownloadMedia(ctx context.Context, ref AppMessageMediaRef) (*AppMessageMedia, error) {
	path := "/" + url.PathEscape(ref.MessageUUID) + "/media/" + url.PathEscape(ref.MediaUUID)
	response, err := send(ctx, c.client.httpClient, c.request(http.MethodGet, path, nil))
	if err != nil {
		return nil, fmt.Errorf("custd: archive image download: %w", err)
	}
	recordMetadata(ctx, response.Header.Get(requestIDHeader))
	body, readErr := io.ReadAll(io.LimitReader(response.Body, maxAppMessageMediaBytes+1))
	if err := errors.Join(readErr, response.Body.Close()); err != nil {
		return nil, fmt.Errorf("custd: read archive image: %w", err)
	}
	if len(body) > maxAppMessageMediaBytes {
		return nil, fmt.Errorf("custd: archive image exceeds %d bytes", maxAppMessageMediaBytes)
	}
	if err := c.client.checkStatus(response.StatusCode, body); err != nil {
		return nil, err
	}
	contentType := response.Header.Get("Content-Type")
	if len(body) == 0 || (contentType != "image/jpeg" && contentType != "image/png") {
		return nil, validationErrorf("custd: archive image response must contain JPEG or PNG data")
	}
	return &AppMessageMedia{ContentType: contentType, Image: body}, nil
}

func (c *AppMessageClient) request(method, path string, body []byte) *bufferedRequest {
	return &bufferedRequest{
		Method: method, URL: strings.TrimRight(c.client.config.BaseURL, "/") + appMessagePath + path,
		Headers: c.client.headers(false), Body: body,
	}
}

func (c *AppMessageClient) jsonRequest(ctx context.Context, method, path string, payload, out any, fields ...string) error {
	body, err := encodeRequestBody(payload, "message archive")
	if err != nil {
		return err
	}
	_, body, err = c.client.exchangeChecked(ctx, "message archive request", c.request(method, path, body))
	if err != nil {
		return err
	}
	return decodeAppMessageResponse(body, out, fields...)
}

func decodeAppMessageResponse(body []byte, out any, fields ...string) error {
	const label = "message archive response"
	document, err := decodeJSONObject(body, label)
	if err != nil {
		return err
	}
	for _, field := range fields {
		if field == "results" || field == "messages" {
			if err := requireJSONObjectList(document, label, field); err != nil {
				return err
			}
		} else if err := requireJSONFields(document, label, field); err != nil {
			return err
		}
	}
	if err := unmarshalResponse(body, out); err != nil {
		return validationWrap("custd: decode "+label, err)
	}
	return nil
}
