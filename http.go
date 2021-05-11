package skadigo

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

// RoundTripper for auto add auth header
// Debug mode will mock out request
type roundTripper struct {
	debug bool
	token string
	r     http.RoundTripper
}

// RoundTrip RoundTripper interface
func (rt roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if rt.debug {
		return &http.Response{
			Request:    r,
			StatusCode: http.StatusNoContent,
		}, nil
	}
	r.Header.Add("Authorization", "Bearer "+rt.token)
	r.Header.Add("Content-Type", "application/json")
	return rt.r.RoundTrip(r)
}

func customRoundTripper(token string, debug bool) http.RoundTripper {
	return roundTripper{
		debug: debug,
		token: token,
		r:     http.DefaultTransport,
	}
}

func (a *Agent) request(r *http.Request) (*http.Response, error) {
	resp, err := a.httpc.Do(r)
	if err != nil {
		return nil, err
	}
	// success
	if resp.StatusCode < 400 {
		return resp, nil
	}
	// failed
	if resp.StatusCode == 401 {
		return nil, errors.New("invalid token")
	}
	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("server error,status code: %d", resp.StatusCode)
	}
	// read error body
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("parse http error resp body failed: %w", err)
	}
	return nil, fmt.Errorf("http request error, status: %d, error: %s", resp.StatusCode, string(body))
}
