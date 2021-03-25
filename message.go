package skadigo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type MessageBody struct {
	Message string `json:"message"`
}

// SendInfo send info level message to your default IM
func (a *Agent) SendInfo(msg string) error {
	body, err := json.Marshal(&MessageBody{msg})
	if err != nil {
		return fmt.Errorf("marshal failed when send info: %w", err)
	}
	req, err := http.NewRequest("POST", a.base+"/agent/info", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request failed when send info: %w", err)
	}
	resp, err := a.httpc.Do(req)
	if err != nil {
		return fmt.Errorf("http failed when send info: %w", err)
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("request failed when send info: %s", resp.Status)
	}
	return nil
}

// SendWarning send warning level message to your default IM
func (a *Agent) SendWarning(msg string) error {
	body, err := json.Marshal(&MessageBody{msg})
	if err != nil {
		return fmt.Errorf("marshal failed when send warning: %w", err)
	}
	req, err := http.NewRequest("POST", a.base+"/agent/warning", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request failed when send warning: %w", err)
	}
	resp, err := a.httpc.Do(req)
	if err != nil {
		return fmt.Errorf("http failed when send warning: %w", err)
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("request failed when send warning: %s", resp.Status)
	}
	return nil
}
