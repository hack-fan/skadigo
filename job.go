package skadigo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type delayedJobInput struct {
	Message  string `json:"message"`
	Duration string `json:"duration,omitempty"`
	// a url, will be called after job status changed,
	// left empty will notify in your default IM,
	// set to "disable" will disable any notify or callback
	Callback string `json:"callback,omitempty"`
}

// AddJobToOther just like what you do in IM, content is <AgentName> <Job>
func (a *Agent) AddJobToOther(content string) error {
	body, err := json.Marshal(&MessageBody{content})
	if err != nil {
		return fmt.Errorf("marshal failed when add job: %w", err)
	}
	req, err := http.NewRequest("POST", a.base+"/agent/job/add", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request failed when add job: %w", err)
	}
	resp, err := a.httpc.Do(req)
	if err != nil {
		return fmt.Errorf("http failed when add job: %w", err)
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("request failed when add job: %s", resp.Status)
	}
	return nil
}

// AddDelayedJob add a delayed job to self.
func (a *Agent) AddDelayedJob(job string, after time.Duration) error {
	return a.addDelayedJob(&delayedJobInput{
		Message:  job,
		Duration: after.String(),
	})
}

// AddDelayedJobSilent add a delayed job to self, and don't report it's status to IM.
func (a *Agent) AddDelayedJobSilent(job string, after time.Duration) error {
	return a.addDelayedJob(&delayedJobInput{
		Message:  job,
		Duration: after.String(),
		Callback: "disable",
	})
}

// AddDelayedJobWithCallback add a delayed job to self, and don't report it's status to IM.
// Skadi will call your callback url with a POST request, body is
// https://github.com/hack-fan/skadi/blob/master/types/job.go#L43
func (a *Agent) AddDelayedJobWithCallback(job string, after time.Duration, callback string) error {
	return a.addDelayedJob(&delayedJobInput{
		Message:  job,
		Duration: after.String(),
		Callback: callback,
	})
}

func (a *Agent) addDelayedJob(dj *delayedJobInput) error {
	body, err := json.Marshal(dj)
	if err != nil {
		return fmt.Errorf("marshal failed when add delayed job: %w", err)
	}
	req, err := http.NewRequest("POST", a.base+"/agent/job/delayed", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request failed when add delayed job: %w", err)
	}
	resp, err := a.httpc.Do(req)
	if err != nil {
		return fmt.Errorf("http failed when add delayed job: %w", err)
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("request failed when add delayed job: %s", resp.Status)
	}
	return nil
}
