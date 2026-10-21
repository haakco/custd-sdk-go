package custd

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
)

// AuthProjectBrandLogo is the stored realm logo metadata the upload route
// returns. UpdatedAt is the server's own timestamp; the image bytes are served
// by the realm and are never part of this answer.
type AuthProjectBrandLogo struct {
	ContentType string `json:"contentType"`
	UpdatedAt   string `json:"updatedAt"`
	ByteSize    int    `json:"byteSize"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

// SetAuthProjectBrandLogo uploads one realm logo as a multipart form with a
// single "logo" part. The server validates the real format from the bytes, so
// the SDK asserts no content type beyond the form boundary.
//
// The answer carries the server's request identifier, because an upload that
// the operator cannot quote is an upload they cannot follow up.
func (c *AuthProjectAdminClient) SetAuthProjectBrandLogo(
	ctx context.Context,
	projectID string,
	environmentID string,
	filename string,
	data []byte,
	options AuthProjectRequestOptions,
) (*AuthProjectBrandLogo, ResponseMetadata, error) {
	body, contentType, err := authProjectBrandLogoBody(filename, data)
	if err != nil {
		return nil, ResponseMetadata{}, err
	}
	var out AuthProjectBrandLogo
	metadata, err := sendAuthProjectRawRequest(
		ctx, c.admin, http.MethodPost,
		authProjectEnvironmentPath(projectID, environmentID)+"/brand/logo",
		body, contentType, &out, authProjectHeaders(options),
	)
	if err != nil {
		return nil, metadata, err
	}
	return &out, metadata, nil
}

// ClearAuthProjectBrandLogo removes a realm's stored logo and its object. The
// answer carries the server's request identifier for the same reason the
// upload does.
func (c *AuthProjectAdminClient) ClearAuthProjectBrandLogo(
	ctx context.Context,
	projectID string,
	environmentID string,
	options AuthProjectRequestOptions,
) (ResponseMetadata, error) {
	return sendAuthProjectRawRequest(
		ctx, c.admin, http.MethodDelete,
		authProjectEnvironmentPath(projectID, environmentID)+"/brand/logo",
		nil, "", nil, authProjectHeaders(options),
	)
}

// authProjectBrandLogoBody packs one image under the "logo" field and returns
// the body with the boundary-bearing content type the route reads.
func authProjectBrandLogoBody(filename string, data []byte) ([]byte, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("logo", filename)
	if err != nil {
		return nil, "", fmt.Errorf("custd: build realm logo upload: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return nil, "", fmt.Errorf("custd: build realm logo upload: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("custd: build realm logo upload: %w", err)
	}
	return body.Bytes(), writer.FormDataContentType(), nil
}

// sendAuthProjectRawRequest sends one auth-project request whose body is
// already encoded, so a multipart upload can carry the bytes the route reads.
//
// It behaves like the admin request helpers the other project-auth methods
// use: the same acting-company and audit labels, the same status handling,
// and the same request identifier in the answer's metadata.
func sendAuthProjectRawRequest(
	ctx context.Context,
	admin *AdminClient,
	method string,
	path string,
	body []byte,
	contentType string,
	out any,
	headers map[string]string,
) (ResponseMetadata, error) {
	path = admin.actingCompanyPath(path)
	merged := admin.withRequestLabels(headers)
	if contentType != "" {
		merged["Content-Type"] = contentType
	}
	if admin.client.config.HTTPClient != nil {
		return sendAuthProjectRawRequestViaDoer(ctx, admin, method, path, body, merged, out)
	}
	return sendAuthProjectRawRequestViaHTTP(ctx, admin, method, path, body, merged, out)
}

// sendAuthProjectRawRequestViaDoer sends the encoded body through the
// configured transport.
func sendAuthProjectRawRequestViaDoer(
	ctx context.Context,
	admin *AdminClient,
	method string,
	path string,
	body []byte,
	headers map[string]string,
	out any,
) (ResponseMetadata, error) {
	requestHeaders := admin.client.headers(false)
	for key, value := range headers {
		requestHeaders[key] = value
	}
	response, err := admin.client.config.HTTPClient.Do(&HTTPRequest{
		Method: method, URL: admin.endpoint(path), Headers: requestHeaders, Body: body,
	})
	if err != nil {
		return ResponseMetadata{}, fmt.Errorf("custd: admin request failed: %w", err)
	}
	metadata := ResponseMetadata{RequestID: responseHeader(response.Headers, requestIDHeader)}
	recordMetadata(ctx, metadata.RequestID)
	if err := admin.client.checkStatus(response.StatusCode, response.Body); err != nil {
		return metadata, err
	}
	return metadata, decodeAdminResponse(response.Body, out)
}

// sendAuthProjectRawRequestViaHTTP sends the encoded body over the built-in
// HTTP client.
func sendAuthProjectRawRequestViaHTTP(
	ctx context.Context,
	admin *AdminClient,
	method string,
	path string,
	body []byte,
	headers map[string]string,
	out any,
) (ResponseMetadata, error) {
	request, err := http.NewRequestWithContext(ctx, method, admin.endpoint(path), bytes.NewReader(body))
	if err != nil {
		return ResponseMetadata{}, fmt.Errorf("custd: create admin request: %w", err)
	}
	for key, value := range admin.client.headers(false) {
		request.Header.Set(key, value)
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response, err := admin.client.httpClient.Do(request)
	if err != nil {
		return ResponseMetadata{}, fmt.Errorf("custd: admin request failed: %w", err)
	}
	responseBody, err := readResponseBody(response.Body)
	if err != nil {
		return ResponseMetadata{}, fmt.Errorf("custd: read admin response: %w", err)
	}
	metadata := ResponseMetadata{RequestID: response.Header.Get(requestIDHeader)}
	recordMetadata(ctx, metadata.RequestID)
	if err := admin.client.checkStatus(response.StatusCode, responseBody); err != nil {
		return metadata, err
	}
	return metadata, decodeAdminResponse(responseBody, out)
}
