package skadigo

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

// NOTICE: Things in this file will be deprecated

// HandlerFunc is your custom handler function
type HandlerFunc func(msg string) (string, error)

// StartWorker NOTICE: this function will be deprecated,
// please use Start() instead
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
