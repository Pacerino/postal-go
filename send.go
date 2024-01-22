package postal

import (
	"context"
	"net/http"
)

const (
	sendBasePath = "/api/v1/send"
	sendPath     = sendBasePath + "/message"
	sendRawPath  = sendBasePath + "/raw"
)

// SendingService is an interface for interfacing with the send
// endpoints of the Postal API.
// See: https://apiv1.postalserver.io/controllers/send.html
type SendingService interface {
	Send(context.Context, *SendRequest) (*SendResponse, *Response, error)
	SendRAW(context.Context, *SendRAWRequest) (*SendResponse, *Response, error)
}

// SendingServiceeOp handles communication with the sending related methods of the Postal API.
type SendingServiceeOp struct {
	client *Client
}

var _ SendingService = &SendingServiceeOp{}

// SendRequest represents a request to the Postal API for a message to send.
type SendRequest struct {
	// The e-mail addresses of the recipients (max 50)
	To []string `json:"to"`
	// The e-mail addresses of any CC contacts (max 50)
	CC []string `json:"cc"`
	// The e-mail addresses of any BCC contacts (max 50)
	BCC []string `json:"bcc"`
	// The e-mail address for the From header
	From string `json:"from"`
	// The e-mail address for the Sender header
	Sender string `json:"sender"`
	// The subject of the e-mail
	Subject string `json:"subject"`
	// The tag of the e-mail
	Tag string `json:"tag"`
	// Set the reply-to address for the mail
	ReplyTo string `json:"reply_to"`
	// The plain text body of the e-mail
	PlainBody string `json:"plain_body"`
	// The HTML body of the e-mail
	HTMLBody string `json:"html_body"`
	// An array of attachments for this e-mail
	Attachments interface{} `json:"attachments"`
	// A hash of additional headers
	Headers map[string]interface{} `json:"headers"`
	// Is this message a bounce?
	Bounce bool `json:"bounce"`
}

// SendRAWRequest represents a request to the Postal API for a message to send.
type SendRAWRequest struct {
	// The address that should be logged as sending the message
	MailFrom string `json:"mail_from"`
	// The addresses this message should be sent to
	RcptTo []string `json:"rcpt_to"`
	// A base64 encoded RFC2822 message to send
	Data string `json:"data"`
	// Is this message a bounce?
	Bounce bool `json:"bounce"`
}

type sendRoot struct {
	Hash *SendResponse `json:"data"`
}

// SendResponse contains all details about a message
type SendResponse struct {
	MessageID string `json:"message_id"`
	Messages  map[string]struct {
		ID    int    `json:"id"`
		Token string `json:"token"`
	} `json:"messages"`
}

// Send a message through the Postal API
func (svc *SendingServiceeOp) Send(ctx context.Context, sendRequest *SendRequest) (*SendResponse, *Response, error) {
	req, err := svc.client.NewRequest(ctx, http.MethodPost, sendPath, sendRequest)
	if err != nil {
		return nil, nil, err
	}
	root := new(sendRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Hash, resp, nil
}

// SendRAW a message through the Postal API using a raw RFC2822 message
func (svc *SendingServiceeOp) SendRAW(ctx context.Context, sendRAWRequest *SendRAWRequest) (*SendResponse, *Response, error) {
	req, err := svc.client.NewRequest(ctx, http.MethodPost, sendRawPath, sendRAWRequest)
	if err != nil {
		return nil, nil, err
	}
	root := new(sendRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Hash, resp, nil
}
