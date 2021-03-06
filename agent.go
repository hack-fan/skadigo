package skadigo

import (
	"bytes"
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
	// optional, job check interval milliseconds, 0 will be default 60000ms(60s)
	Interval int
	// can be many kind of logger, look up the def
	Logger Logger
}

// Agent or client
type Agent struct {
	// base url
	base     string
	handle   HandlerFunc
	httpc    *http.Client
	interval time.Duration
	log      Logger
}

// New skadi agent instance, you can Start() it later.
func New(token, server string, handler HandlerFunc, opts *Options) *Agent {
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
		if opts.Interval > 0 {
			interval = time.Duration(opts.Interval) * time.Millisecond
		}
		if opts.Logger != nil {
			log = opts.Logger
		}
	}
	if log == nil {
		log = defaultLogger{}
	}
	return &Agent{
		base:   server,
		handle: handler,
		httpc: &http.Client{
			Transport: customRoundTripper(token),
			Timeout:   timeout,
		},
		interval: interval,
		log:      log,
	}
}

// Start the agent service,blocked,check job and run it in endless loop.
func (a *Agent) Start() {
	for range time.Tick(a.interval) {
		go a.pullJobAndRun()
	}
}

// async run in loop
func (a *Agent) pullJobAndRun() {
	resp, err := a.httpc.Get(a.base + "/agent/job")
	if err != nil {
		// TODO: log error
		return
	}
	// no job
	if resp.StatusCode == 204 {
		return
	}
	if resp.StatusCode != 200 {
		// log other errors
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		a.log.Errorf("read body failed when pull job: %s", err)
		return
	}
	defer resp.Body.Close()
	var job = new(JobBasic)
	err = json.Unmarshal(body, job)
	if err != nil {
		a.log.Errorf("invalid body struct when pull job: %s", err)
		return
	}
	result, err := a.handle(job.Message)
	if err != nil {
		a.fail(job.ID, result)
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
