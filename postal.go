package postal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	libraryVersion = "0.1.0"
	userAgent      = "gopostal/" + libraryVersion
	mediaType      = "application/json"

	statusSuccess        = "success"
	statusParameterError = "parameter-error"
	statusError          = "error"
)

// Client manages communication with the Postal API.
type Client struct {
	// HTTP Client used to communicate with the Postal API.
	HTTPClient *http.Client

	// Base URL for API requests.
	BaseURL *url.URL

	// ApiKey used for authentication with the Postal API.
	ApiKey string

	// Optional extra HTTP headers to set on every request to the API.
	headers map[string]string

	// User agent for client
	UserAgent string

	// Optional function called after every successful request made to the Postal API
	onRequestCompleted RequestCompletionCallback

	// Services used for talking to different parts of the Postal API.
	Messages MessagesService
	Send     SendingService
}

// RequestCompletionCallback defines the type of the request callback function
type RequestCompletionCallback func(*http.Request, *http.Response)

// Response is a Postal API response. This wraps the standard http.Response returned from Postal.
type Response struct {
	// HTTP response that caused this response
	*http.Response

	// API response status
	Status string `json:"status"`

	// Response time
	Time float32 `json:"time"`

	// Additional response flags
	Flags map[string]interface{} `json:"flags"`

	// Response data
	Data interface{} `json:"data"`
}

// An ErrorResponse reports the error caused by an API request
type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	// Error data
	Data interface{} `json:"data"`
}

// NewClient returns a new Postal API client, using the given
// resty.Client to perform all requests.
func NewClient(baseURL string, akey string) *Client {
	httpClient := http.DefaultClient
	burl, _ := url.Parse(baseURL)

	c := &Client{HTTPClient: httpClient, UserAgent: userAgent, BaseURL: burl, ApiKey: akey}

	c.Messages = &MessagesServiceOp{client: c}
	c.Send = &SendingServiceeOp{client: c}

	return c
}

// OnRequestCompleted sets the Postal API request completion callback
func (c *Client) OnRequestCompleted(rc RequestCompletionCallback) {
	c.onRequestCompleted = rc
}

// newResponse creates a new Response for the provided http.Response
func newResponse(r *http.Response) *Response {
	response := Response{Response: r}
	return &response
}

// SetBaseURL is a client option for setting the base URL.
func (c *Client) SetBaseURL(bu string) {
	u, _ := url.Parse(bu)
	c.BaseURL = u
}

// SetBaseURL is a client option for setting the base URL.
func (c *Client) SetApiKey(akey string) {
	c.ApiKey = akey
}

// NewRequest creates an API request. A relative URL can be provided in urlStr, which will be resolved to the
// BaseURL of the Client. Relative URLS should always be specified without a preceding slash. If specified, the
// value pointed to by body is JSON encoded and included in as the request body.
func (c *Client) NewRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var req *http.Request
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		req, err = http.NewRequest(method, u.String(), nil)
		if err != nil {
			return nil, err
		}

	default:
		buf := new(bytes.Buffer)
		if body != nil {
			err = json.NewEncoder(buf).Encode(body)
			if err != nil {
				return nil, err
			}
		}

		req, err = http.NewRequest(method, u.String(), buf)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", mediaType)
	}

	for k, v := range c.headers {
		req.Header.Add(k, v)
	}

	req.Header.Set("X-Server-API-Key", c.ApiKey)
	req.Header.Set("User-Agent", c.UserAgent)

	return req, nil
}

// Do sends an API request and returns the API response. The API response is JSON decoded and stored in the value
// pointed to by v, or returned as an error if an API error has occurred.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	resp, err := DoRequestWithClient(ctx, c.HTTPClient, req)
	if err != nil {
		return nil, err
	}
	if c.onRequestCompleted != nil {
		c.onRequestCompleted(req, resp)
	}

	defer func() {
		// Ensure the response body is fully read and closed
		// before we reconnect, so that we reuse the same TCPConnection.
		// Close the previous response's body. But read at least some of
		// the body so if it's small the underlying TCP connection will be
		// re-used. No need to check for errors: if it fails, the Transport
		// won't reuse it anyway.
		const maxBodySlurpSize = 2 << 10
		if resp.ContentLength == -1 || resp.ContentLength <= maxBodySlurpSize {
			io.CopyN(io.Discard, resp.Body, maxBodySlurpSize)
		}

		if rerr := resp.Body.Close(); err == nil {
			err = rerr
		}
	}()

	response := newResponse(resp)
	body, err := CheckResponse(resp)
	if err != nil {
		return response, err
	}

	if resp.StatusCode != http.StatusNoContent && v != nil {
		buf := bytes.NewBuffer(body)
		err = json.NewDecoder(buf).Decode(v)
		if err != nil {
			return nil, err
		}
	}

	return response, err
}

// DoRequest submits an HTTP request.
func DoRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return DoRequestWithClient(ctx, http.DefaultClient, req)
}

// DoRequestWithClient submits an HTTP request using the specified client.
func DoRequestWithClient(
	ctx context.Context,
	client *http.Client,
	req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return client.Do(req)
}

// ErrorResponse implements the error interface.
func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Data)
}

func CheckResponse(r *http.Response) ([]byte, error) {
	var response Response
	errorResponse := &ErrorResponse{Response: r}
	data, err := io.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		err = json.Unmarshal(data, &response)
		if err != nil {
			errorResponse.Data = err.Error()
			return nil, errorResponse
		}
		if response.Status != statusSuccess {
			errorResponse.Data = response.Data.(map[string]interface{})["message"]
			return nil, errorResponse
		}
	}
	return data, nil
}
