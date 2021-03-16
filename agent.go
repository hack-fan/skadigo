package skadigo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// Logger can be logrus, zap, etc...
type Logger interface {
	Errorf(string, ...interface{})
}

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

// Options for agent
type Options struct {
	// optional, you can custom http client timeout milliseconds, default value is 3000ms(3s)
	Timeout int
	// can be many kind of logger, look up the def
	Logger Logger
}

// Agent or client
type Agent struct {
	base     string
	httpc    *http.Client
	interval time.Duration
	log      Logger
}

// New skadi agent instance, you can just use it for sending messages,
// or StartWorker to handle your command later.
func New(token, server string, opts *Options) *Agent {
	// check options
	base, err := url.Parse(server)
	if err != nil || base.Host == "" {
		panic("invalid skadi server address")
	}
	var timeout = 3 * time.Second
	var interval = time.Minute
	var log Logger
	if opts != nil {
		if opts.Timeout > 0 {
			timeout = time.Duration(opts.Timeout) * time.Millisecond
		}
		if opts.Logger != nil {
			log = opts.Logger
		}
	}
	if log == nil {
		log = defaultLogger{}
	}
	return &Agent{
		base: server,
		httpc: &http.Client{
			Transport: customRoundTripper(token),
			Timeout:   timeout,
		},
		interval: interval,
		log:      log,
	}
}

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

// RoundTripper for auto add auth header
type roundTripper struct {
	token string
	r     http.RoundTripper
}

// RoundTripper interface
func (rt roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", "Bearer "+rt.token)
	r.Header.Add("Content-Type", "application/json")
	return rt.r.RoundTrip(r)
}

func customRoundTripper(token string) http.RoundTripper {
	return roundTripper{
		token: token,
		r:     http.DefaultTransport,
	}
}

type defaultLogger struct{}

func (defaultLogger) Errorf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
