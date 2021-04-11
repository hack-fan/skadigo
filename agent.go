package skadigo

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Logger can be logrus, zap, etc...
type Logger interface {
	Errorf(string, ...interface{})
}

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
// token: your agent token
// server: skadi agent api server, https://github.com/hack-fan/skadi
// nolint
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

// RoundTripper for auto add auth header
type roundTripper struct {
	token string
	r     http.RoundTripper
}

// RoundTrip RoundTripper interface
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
