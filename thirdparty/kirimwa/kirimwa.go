package kirimwa

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/mhdiiilham/gosm/logger"
)

// Client provides parameters required to use KirimWA.id API.
type Client struct {
	apiKey     string
	devideID   string
	baseURL    string
	httpClient *http.Client
}

// PostMessageRequest represent the payload required for sending a whatsapp message.
type PostMessageRequest struct {
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
	DeviceID    string `json:"device_id"`
	MessageType string `json:"text"`
}

// PostMessageResponse represent the success response when sending a whatsapp message.
type PostMessageResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// NewKirimWAClient initializes a new KirimWA.id client instance.
func NewKirimWAClient(apiKey string, deviceID string) *Client {
	return &Client{
		apiKey:     apiKey,
		devideID:   deviceID,
		baseURL:    "https://api.kirimwa.id/v1/",
		httpClient: http.DefaultClient,
	}
}

// SendMessage sent WhatsApp `message` to `destination`.
func (c *Client) SendMessage(ctx context.Context, destination string, message string) (id, status string, err error) {
	messageRequestBody := PostMessageRequest{
		PhoneNumber: destination,
		Message:     message,
		DeviceID:    c.devideID,
		MessageType: "text",
	}

	var payload bytes.Buffer
	if err := json.NewEncoder(&payload).Encode(messageRequestBody); err != nil {
		return "", "", err
	}

	var response PostMessageResponse
	if err := c.request(ctx, http.MethodPost, "messages", &payload, nil, &response); err != nil {
		logger.Errorf(ctx, "SendMessage", "failed from request: %v", err)
		return "", "", err
	}

	return response.Status, response.Status, nil
}

func (c *Client) request(ctx context.Context, method string, urlPath string, body io.Reader, headers map[string]string, responseDst any) error {
	request, err := http.NewRequest(method, fmt.Sprintf("%s/%s", c.baseURL, urlPath), body)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.apiKey)

	// set all the headers
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		logger.Errorf(ctx, "KirimWAClient.request", "c.httpClient.Do return err: %v", err)
		return err
	}

	// Other than created mark as failed
	if response.StatusCode != http.StatusCreated {
		logger.Errorf(ctx, "KirimWAClient.request", "failed to post message: %v", err)
		return errors.New("INTERNAL_SERVER_ERROR")
	}

	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(&responseDst); err != nil {
		return err
	}

	return nil
}
