package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Token:      token,
		HTTPClient: &http.Client{},
	}
}

type APIError struct {
	Error string `json:"error"`
}

func (c *Client) DoRequest(method, path string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshaling request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

func CheckError(body []byte, statusCode int) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}

	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error != "" {
		return fmt.Errorf("API error (%d): %s", statusCode, apiErr.Error)
	}

	return fmt.Errorf("API error (%d): %s", statusCode, string(body))
}

// Login authenticates and returns user info + token (no auth header needed)
func (c *Client) Login(email, password string) ([]byte, error) {
	body, status, err := c.DoRequest("POST", "/api/v1/session", map[string]string{
		"email_address": email,
		"password":      password,
	})
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

// Join creates account via join code
func (c *Client) Join(joinCode, name, email, password string) ([]byte, error) {
	body, status, err := c.DoRequest("POST", "/api/v1/join", map[string]string{
		"join_code":     joinCode,
		"name":          name,
		"email_address": email,
		"password":      password,
	})
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

// FirstRun sets up the initial account
func (c *Client) FirstRun(name, email, password string) ([]byte, error) {
	body, status, err := c.DoRequest("POST", "/api/v1/first_run", map[string]string{
		"name":           name,
		"email_address":  email,
		"password":       password,
	})
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

// Users
func (c *Client) ListUsers() ([]byte, error) {
	body, status, err := c.DoRequest("GET", "/api/v1/users", nil)
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

func (c *Client) CreateUser(params map[string]interface{}) ([]byte, error) {
	body, status, err := c.DoRequest("POST", "/api/v1/users", params)
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

// Rooms
func (c *Client) ListRooms() ([]byte, error) {
	body, status, err := c.DoRequest("GET", "/api/v1/rooms", nil)
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

func (c *Client) ShowRoom(id string) ([]byte, error) {
	body, status, err := c.DoRequest("GET", "/api/v1/rooms/"+id, nil)
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

func (c *Client) CreateRoom(params map[string]interface{}) ([]byte, error) {
	body, status, err := c.DoRequest("POST", "/api/v1/rooms", params)
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

func (c *Client) UpdateRoom(id string, params map[string]interface{}) ([]byte, error) {
	body, status, err := c.DoRequest("PATCH", "/api/v1/rooms/"+id, params)
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

func (c *Client) DeleteRoom(id string) error {
	body, status, err := c.DoRequest("DELETE", "/api/v1/rooms/"+id, nil)
	if err != nil {
		return err
	}
	if err := CheckError(body, status); err != nil {
		return err
	}
	return nil
}

func (c *Client) DirectRoom(userID string) ([]byte, error) {
	body, status, err := c.DoRequest("POST", "/api/v1/rooms/direct", map[string]interface{}{
		"user_id": userID,
	})
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

// Messages
func (c *Client) ListMessages(roomID string, params map[string]string) ([]byte, error) {
	path := "/api/v1/rooms/" + roomID + "/messages"
	query := ""
	for k, v := range params {
		if v != "" {
			if query == "" {
				query = "?"
			} else {
				query += "&"
			}
			query += k + "=" + v
		}
	}
	body, status, err := c.DoRequest("GET", path+query, nil)
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

func (c *Client) CreateMessage(roomID, msgBody string) ([]byte, error) {
	body, status, err := c.DoRequest("POST", "/api/v1/rooms/"+roomID+"/messages", map[string]string{
		"body": msgBody,
	})
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

func (c *Client) DeleteMessage(roomID, messageID string) error {
	body, status, err := c.DoRequest("DELETE", "/api/v1/rooms/"+roomID+"/messages/"+messageID, nil)
	if err != nil {
		return err
	}
	if err := CheckError(body, status); err != nil {
		return err
	}
	return nil
}

// Boosts
func (c *Client) CreateBoost(messageID, content string) ([]byte, error) {
	body, status, err := c.DoRequest("POST", "/api/v1/messages/"+messageID+"/boosts", map[string]string{
		"content": content,
	})
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

func (c *Client) DeleteBoost(messageID, boostID string) error {
	body, status, err := c.DoRequest("DELETE", "/api/v1/messages/"+messageID+"/boosts/"+boostID, nil)
	if err != nil {
		return err
	}
	if err := CheckError(body, status); err != nil {
		return err
	}
	return nil
}

// Search
func (c *Client) Search(query string, params map[string]string) ([]byte, error) {
	path := "/api/v1/search?q=" + query
	for k, v := range params {
		if v != "" {
			path += "&" + k + "=" + v
		}
	}
	body, status, err := c.DoRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

// Presence
func (c *Client) ListPresence() ([]byte, error) {
	body, status, err := c.DoRequest("GET", "/api/v1/users/presence", nil)
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

// Involvement
func (c *Client) UpdateInvolvement(roomID, level string) ([]byte, error) {
	body, status, err := c.DoRequest("PATCH", "/api/v1/rooms/"+roomID+"/involvement", map[string]string{
		"involvement": level,
	})
	if err != nil {
		return nil, err
	}
	if err := CheckError(body, status); err != nil {
		return nil, err
	}
	return body, nil
}

// Watch opens a WebSocket connection to the ActionCable API channel and calls
// onMessage for each received event. It blocks until the connection is closed
// or an error occurs.
func (c *Client) Watch(onMessage func([]byte)) error {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("parsing base URL: %w", err)
	}

	scheme := "ws"
	if u.Scheme == "https" {
		scheme = "wss"
	}
	u.Scheme = scheme
	u.Path = "/cable"
	q := u.Query()
	q.Set("token", c.Token)
	u.RawQuery = q.Encode()

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	header := http.Header{}
	header.Set("Origin", c.BaseURL)
	conn, resp, err := dialer.Dial(u.String(), header)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("connecting to WebSocket: %w (HTTP %d)", err, resp.StatusCode)
		}
		return fmt.Errorf("connecting to WebSocket: %w", err)
	}
	defer conn.Close()

	// Subscribe to the ApiChannel
	subscribe := map[string]interface{}{
		"command":    "subscribe",
		"identifier": `{"channel":"ApiChannel"}`,
	}
	if err := conn.WriteJSON(subscribe); err != nil {
		return fmt.Errorf("subscribing: %w", err)
	}

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("reading message: %w", err)
		}

		var frame struct {
			Type       string          `json:"type"`
			Message    json.RawMessage `json:"message"`
			Identifier string          `json:"identifier"`
		}
		if err := json.Unmarshal(raw, &frame); err != nil {
			continue
		}

		// Skip ActionCable internal frames (welcome, ping, confirm_subscription)
		if frame.Type != "" {
			continue
		}

		// Only deliver messages from our channel
		if frame.Identifier != `{"channel":"ApiChannel"}` {
			continue
		}

		if frame.Message != nil {
			onMessage([]byte(frame.Message))
		}
	}
}
