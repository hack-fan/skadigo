package skadigo

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
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

// Options for agent
type Options struct {
	// optional, you can custom http client timeout milliseconds, default value is 3000ms(3s)
	Timeout int
	// optional, job check interval milliseconds, 0 will be default 60000ms(60s)
	Interval int
}

// Agent or client
type Agent struct {
	// base url
	base     string
	handle   HandlerFunc
	httpc    *http.Client
	interval time.Duration
}

func New(token, server string, handler HandlerFunc, opts *Options) *Agent {
	// check options
	base, err := url.Parse(server)
	if err != nil || base.Host == "" {
		panic("invalid skadi server address")
	}
	var timeout = 3 * time.Second
	var interval = time.Minute
	if opts != nil {
		if opts.Timeout > 0 {
			timeout = time.Duration(opts.Timeout) * time.Millisecond
		}
		if opts.Interval > 0 {
			interval = time.Duration(opts.Interval) * time.Millisecond
		}
	}
	return &Agent{
		base:   server,
		handle: handler,
		httpc: &http.Client{
			Transport: customRoundTripper(token),
			Timeout:   timeout,
		},
		interval: interval,
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
		// TODO: log error
		return
	}
	defer resp.Body.Close()
	var job = new(JobBasic)
	err = json.Unmarshal(body, job)
	if err != nil {
		// TODO: log error
		return
	}
	result, err := a.handle(job.Message)
	if err != nil {
		a.fail(job.ID, result)
	}
	a.succeed(job.ID, result)
}

func (a *Agent) succeed(id, result string) {
	req, err := http.NewRequest("PUT", a.base+"/agent/jobs/"+id+"/succeed", nil)
	if err != nil {
		// TODO: log error
		return
	}
	resp, err := a.httpc.Do(req)
	if err != nil {
		// TODO: log error
		return
	}
	if resp.StatusCode != 204 {
		// TODO: log error
		return
	}
}

func (a *Agent) fail(id, result string) {
	req, err := http.NewRequest("PUT", a.base+"/agent/jobs/"+id+"/fail", nil)
	if err != nil {
		// TODO: log error
		return
	}
	resp, err := a.httpc.Do(req)
	if err != nil {
		// TODO: log error
		return
	}
	if resp.StatusCode != 204 {
		// TODO: log error
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
	return rt.r.RoundTrip(r)
}

func customRoundTripper(token string) http.RoundTripper {
	return roundTripper{
		token: token,
		r:     http.DefaultTransport,
	}
}
