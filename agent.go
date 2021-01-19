package skadigo

import (
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

// Options for agent
type Options struct {
	// required, a base url,like https://agent.letserver.run or localhost:8080
	Server string
	// required, you must supply a func for your job processing
	Handler HandlerFunc
	// optional, you can custom http client
	HTTPClient *http.Client
	// optional, job check interval milliseconds, 0 will be default 60000ms/60s/1min
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

func New(opts *Options) *Agent {
	// check options
	base := opts.Server
	var httpc *http.Client
	if opts.HTTPClient != nil {
		httpc = opts.HTTPClient
	} else {
		httpc = &http.Client{
			Timeout: 3 * time.Second,
		}
	}
	var interval time.Duration
	if opts.Interval > 0 {
		interval = time.Duration(opts.Interval) * time.Millisecond
	} else {
		interval = time.Minute
	}
	return &Agent{
		base:     base,
		handle:   opts.Handler,
		httpc:    httpc,
		interval: interval,
	}
}
