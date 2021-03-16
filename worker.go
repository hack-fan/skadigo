package skadigo

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

// JobBasic is redefined the job struct for zero deps
type JobBasic struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// JobResult is redefine the result struct for zero deps
type JobResult struct {
	Result string `json:"result"`
}

// HandlerFunc is your custom handler function
type HandlerFunc func(msg string) (string, error)

// StartWorker will start the agent worker service,blocked,check job and run it in endless loop.
// please start it only once per agent, avoid exceed the api limit.
// ctx: you can use cancelable context for gracefully shutdown worker, or just use context.Background
// handler: required, handle command message and return result or error
// interval: optional, job check interval milliseconds, 0 will be default 60000ms(60s)
func (a *Agent) StartWorker(ctx context.Context, handler HandlerFunc, interval time.Duration) {
	if interval == 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case <-ticker.C:
			a.pullJobAndRun(handler)
		}
	}
}

// async run in loop
func (a *Agent) pullJobAndRun(handler HandlerFunc) {
	resp, err := a.httpc.Get(a.base + "/agent/job")
	if err != nil {
		a.log.Errorf("pull job failed: %s", err)
		return
	}
	// no job
	if resp.StatusCode == 204 {
		return
	}
	// other status
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		a.log.Errorf("read body failed when pull job: %s", err)
		return
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		a.log.Errorf("unknown server error: %s %s", http.StatusText(resp.StatusCode), string(body))
		return
	}
	// job got
	var job = new(JobBasic)
	err = json.Unmarshal(body, job)
	if err != nil {
		a.log.Errorf("invalid body struct when pull job: %s", err)
		return
	}
	result, err := handler(job.Message)
	if err != nil {
		a.fail(job.ID, err.Error())
		return
	}
	a.succeed(job.ID, result)
}

func (a *Agent) succeed(id, result string) {
	body, err := json.Marshal(&JobResult{result})
	if err != nil {
		a.log.Errorf("invalid result when report job succeeded: %s", err)
		return
	}
	req, err := http.NewRequest("PUT", a.base+"/agent/jobs/"+id+"/succeed", bytes.NewReader(body))
	if err != nil {
		a.log.Errorf("invalid request when report job succeeded: %s", err)
		return
	}
	resp, err := a.httpc.Do(req)
	if err != nil {
		a.log.Errorf("request failed when report job succeeded: %s", err)
		return
	}
	if resp.StatusCode != 204 {
		a.log.Errorf("request failed status when report job succeeded: %s", err)
		return
	}
}

func (a *Agent) fail(id, result string) {
	body, err := json.Marshal(&JobResult{result})
	if err != nil {
		a.log.Errorf("invalid result when report job failed: %s", err)
		return
	}
	req, err := http.NewRequest("PUT", a.base+"/agent/jobs/"+id+"/fail", bytes.NewReader(body))
	if err != nil {
		a.log.Errorf("invalid request when report job failed: %s", err)
		return
	}
	resp, err := a.httpc.Do(req)
	if err != nil {
		a.log.Errorf("request failed when report job failed: %s", err)
		return
	}
	if resp.StatusCode != 204 {
		a.log.Errorf("request failed status when report job failed: %s", err)
		return
	}
}
