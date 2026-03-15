package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
func (c *Client) Search(query string, limit string) ([]byte, error) {
	path := "/api/v1/search?q=" + query
	if limit != "" {
		path += "&limit=" + limit
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
