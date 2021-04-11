package skadigo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

// Handler is your message processor.
// Please log your handler result and errors yourself.
// You can use id to send running status,
// or just omit the id param.
type Handler func(id, msg string) (string, error)

// Start will start the agent worker service,blocked,check job and run it in endless loop.
// please start it only once per agent, avoid exceed the api limit.
// ctx: you can use cancelable context for gracefully shutdown worker, or just use context.Background
// handler: required, handle command message and return result or error
// interval: optional, job check interval milliseconds, 0 will be default 60000ms(60s)
// return value will be nil when context canceled, otherwise there will be an error.
func (a *Agent) Start(ctx context.Context, handler Handler, interval time.Duration) error {
	if interval == 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// only 401 will return an error
			job, ok, err := a.check()
			if err != nil {
				a.log.Errorf("skadi agent worker will be exit with error: %s", err)
				return err
			}
			if ok {
				a.run(job, handler)
			}
		}
	}
}

func (a *Agent) check() (*JobBasic, bool, error) {
	resp, err := a.httpc.Get(a.base + "/agent/job")
	if err != nil {
		a.log.Errorf("pull job failed: %s", err)
		// network error will be ignored for retrying
		return nil, false, nil
	}
	// no job
	if resp.StatusCode == 204 {
		return nil, false, nil
	}
	// error token, this error will be returned, then worker exit with error
	if resp.StatusCode == 401 {
		return nil, false, errors.New("invalid token")
	}
	// other unique errors
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		a.log.Errorf("read body failed when pull job: %s", err)
		return nil, false, nil
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		a.log.Errorf("unknown server error: %s %s", http.StatusText(resp.StatusCode), string(body))
		return nil, false, nil
	}
	// job got
	var job = new(JobBasic)
	err = json.Unmarshal(body, job)
	if err != nil {
		// error struct, this error will be returned, then worker exit with error
		return nil, false, errors.New("invalid job body, please upgrade your agent")
	}
	return job, true, nil
}

func (a *Agent) run(job *JobBasic, handler Handler) {
	result, err := handler(job.ID, job.Message)
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
