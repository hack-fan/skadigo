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
	req, err := http.NewRequest("POST", a.base+"/agent/message/info", bytes.NewReader(body))
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
	req, err := http.NewRequest("POST", a.base+"/agent/message/warning", bytes.NewReader(body))
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

// SendText send text message to your default IM
// Warning: If default IM is wechat MP, the message may be discarded when you are not active.
func (a *Agent) SendText(msg string) error {
	body, err := json.Marshal(&MessageBody{msg})
	if err != nil {
		return fmt.Errorf("marshal failed when send text: %w", err)
	}
	req, err := http.NewRequest("POST", a.base+"/agent/message/text", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request failed when send text: %w", err)
	}
	resp, err := a.httpc.Do(req)
	if err != nil {
		return fmt.Errorf("http failed when send text: %w", err)
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("request failed when send text: %s", resp.Status)
	}
	return nil
}

// SendAuto send text message to your default IM
// If default IM is wechat MP, it will try kf text message first, if failed, try template message.
func (a *Agent) SendAuto(msg string) error {
	body, err := json.Marshal(&MessageBody{msg})
	if err != nil {
		return fmt.Errorf("marshal failed when send auto: %w", err)
	}
	req, err := http.NewRequest("POST", a.base+"/agent/message/auto", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request failed when send auto: %w", err)
	}
	resp, err := a.httpc.Do(req)
	if err != nil {
		return fmt.Errorf("http failed when send auto: %w", err)
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("request failed when send auto: %s", resp.Status)
	}
	return nil
}

func (a *Agent) Info(args ...interface{}) {
	err := a.SendInfo(fmt.Sprint(args...))
	if err != nil {
		a.log.Errorf("skadigo send info failed: %s", err)
	}
}

func (a *Agent) Infof(format string, args ...interface{}) {
	err := a.SendInfo(fmt.Sprintf(format, args...))
	if err != nil {
		a.log.Errorf("skadigo send info failed: %s", err)
	}
}

func (a *Agent) Warn(args ...interface{}) {
	err := a.SendWarning(fmt.Sprint(args...))
	if err != nil {
		a.log.Errorf("skadigo send info failed: %s", err)
	}
}

func (a *Agent) Warnf(format string, args ...interface{}) {
	err := a.SendWarning(fmt.Sprintf(format, args...))
	if err != nil {
		a.log.Errorf("skadigo send info failed: %s", err)
	}
}
