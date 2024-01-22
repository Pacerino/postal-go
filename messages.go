package postal

import (
	"context"
	"net/http"
)

const (
	messagesBasePath = "/api/v1/messages"
	detailsPath      = messagesBasePath + "/message"
	deliveriesPath   = messagesBasePath + "/deliveries"
)

// MessagesService is an interface for interfacing with the message
// endpoints of the Postal API.
// See: https://apiv1.postalserver.io/controllers/messages.html
type MessagesService interface {
	GetMessage(context.Context, *GetMessageRequest) (*MessageDetails, *Response, error)
	GetDeliveries(context.Context, *GetDeliveriesRequest) (*[]MessageDeliveries, *Response, error)
}

// MessagesServiceOp handles communication with the message related methods of the Postal API.
type MessagesServiceOp struct {
	client *Client
}

var _ MessagesService = &MessagesServiceOp{}

// GetMessageRequest represents a request to the Postal API for a message.
type GetMessageRequest struct {
	// ID of the message
	ID int `json:"id"`
	// Expansions is either a string array or a bool
	Expansions interface{} `json:"_expansions"`
}

// GetDeliveriesRequest represents a request to the Postal API for a message.
type GetDeliveriesRequest struct {
	ID int `json:"id"`
}

type messageRoot struct {
	Message *MessageDetails `json:"data"`
}

// MessageDetails contains all details about a message
type MessageDetails struct {
	ID     int    `json:"id"`
	Token  string `json:"token"`
	Status struct {
		Status              string      `json:"status"`
		LastDeliveryAttempt float64     `json:"last_delivery_attempt"`
		Held                bool        `json:"held"`
		HoldExpiry          interface{} `json:"hold_expiry"`
	} `json:"status"`
	Details struct {
		RcptTo          string      `json:"rcpt_to"`
		MailFrom        string      `json:"mail_from"`
		Subject         string      `json:"subject"`
		MessageID       string      `json:"message_id"`
		Timestamp       float64     `json:"timestamp"`
		Direction       string      `json:"direction"`
		Size            interface{} `json:"size"`
		Bounce          bool        `json:"bounce"`
		BounceForID     int         `json:"bounce_for_id"`
		Tag             interface{} `json:"tag"`
		ReceivedWithSsl interface{} `json:"received_with_ssl"`
	} `json:"details"`
	Inspection struct {
		Inspected     bool        `json:"inspected"`
		Spam          bool        `json:"spam"`
		SpamScore     float64     `json:"spam_score"`
		Threat        bool        `json:"threat"`
		ThreatDetails interface{} `json:"threat_details"`
	} `json:"inspection"`
	PlainBody   interface{}   `json:"plain_body"`
	HTMLBody    interface{}   `json:"html_body"`
	Attachments []interface{} `json:"attachments"`
	Headers     struct {
	} `json:"headers"`
	RawMessage      string `json:"raw_message"`
	ActivityEntries struct {
		Loads  []interface{} `json:"loads"`
		Clicks []interface{} `json:"clicks"`
	} `json:"activity_entries"`
}

type deliveriesRoot struct {
	Deliveries *[]MessageDeliveries `json:"data"`
}

// MessageDeliveries contains an array of deliveries which have been attempted for this message
type MessageDeliveries struct {
	ID          int     `json:"id"`
	Status      string  `json:"status"`
	Details     string  `json:"details"`
	Output      string  `json:"output"`
	SentWithSsl bool    `json:"sent_with_ssl"`
	LogID       string  `json:"log_id"`
	Time        float64 `json:"time"`
	Timestamp   float64 `json:"timestamp"`
}

// GetMessage returns all details about a message
func (mvc *MessagesServiceOp) GetMessage(ctx context.Context, getRequest *GetMessageRequest) (*MessageDetails, *Response, error) {
	req, err := mvc.client.NewRequest(ctx, http.MethodPost, detailsPath, getRequest)
	if err != nil {
		return nil, nil, err
	}
	root := new(messageRoot)
	resp, err := mvc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Message, resp, nil
}

// GetDeliveries returns an array of deliveries which have been attempted for this message
func (mvc *MessagesServiceOp) GetDeliveries(ctx context.Context, getRequest *GetDeliveriesRequest) (*[]MessageDeliveries, *Response, error) {

	req, err := mvc.client.NewRequest(ctx, http.MethodPost, deliveriesPath, getRequest)
	if err != nil {
		return nil, nil, err
	}
	root := new(deliveriesRoot)
	resp, err := mvc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Deliveries, resp, nil
}
