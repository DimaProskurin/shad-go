//go:build !solution

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Rules []ConfigRule `yaml:"rules"`
}

type ConfigRule struct {
	Endpoint                string   `yaml:"endpoint"`
	ForbiddenUserAgents     []string `yaml:"forbidden_user_agents"`
	ForbiddenHeaders        []string `yaml:"forbidden_headers"`
	RequiredHeaders         []string `yaml:"required_headers"`
	MaxRequestLengthBytes   int64    `yaml:"max_request_length_bytes"`
	MaxResponseLengthBytes  int64    `yaml:"max_response_length_bytes"`
	ForbiddenResponseCodes  []int    `yaml:"forbidden_response_codes"`
	ForbiddenRequestRegexp  []string `yaml:"forbidden_request_re"`
	ForbiddenResponseRegexp []string `yaml:"forbidden_response_re"`
}

type Firewall struct {
	endpointRules map[string]Rule
}

type Rule struct {
	ForbiddenUserAgents     []*regexp.Regexp
	ForbiddenHeaders        []ForbiddenHeader
	RequiredHeaders         []string
	MaxRequestLengthBytes   int64
	MaxResponseLengthBytes  int64
	ForbiddenResponseCodes  []int
	ForbiddenRequestRegexp  []*regexp.Regexp
	ForbiddenResponseRegexp []*regexp.Regexp
}

type ForbiddenHeader struct {
	Key   string
	Value string
}

func New(conf *Config) (*Firewall, error) {
	if conf == nil {
		return nil, errors.New("firewall config is nil")
	}

	endpointRules := make(map[string]Rule)
	for _, r := range conf.Rules {
		fbdReq, err := CompileSliceRegexp(r.ForbiddenRequestRegexp)
		if err != nil {
			return nil, err
		}

		fbdResp, err := CompileSliceRegexp(r.ForbiddenResponseRegexp)
		if err != nil {
			return nil, err
		}

		fbdUA, err := CompileSliceRegexp(r.ForbiddenUserAgents)
		if err != nil {
			return nil, err
		}

		fbdHeaders := make([]ForbiddenHeader, 0, len(r.ForbiddenHeaders))
		for _, s := range r.ForbiddenHeaders {
			parts := strings.SplitN(s, ": ", 2)
			if len(parts) != 2 {
				return nil, errors.New("wrong header: " + s)
			}
			fbdHeaders = append(fbdHeaders, ForbiddenHeader{Key: parts[0], Value: parts[1]})
		}

		endpointRules[r.Endpoint] = Rule{
			ForbiddenUserAgents:     fbdUA,
			ForbiddenHeaders:        fbdHeaders,
			RequiredHeaders:         r.RequiredHeaders,
			MaxRequestLengthBytes:   r.MaxRequestLengthBytes,
			MaxResponseLengthBytes:  r.MaxResponseLengthBytes,
			ForbiddenResponseCodes:  r.ForbiddenResponseCodes,
			ForbiddenRequestRegexp:  fbdReq,
			ForbiddenResponseRegexp: fbdResp,
		}
	}
	return &Firewall{endpointRules: endpointRules}, nil
}

func (f *Firewall) CheckRequest(endpoint string, req *http.Request) (ok bool) {
	rule := f.endpointRules[endpoint]

	for _, h := range rule.RequiredHeaders {
		if len(req.Header.Get(h)) == 0 {
			return false
		}
	}

	for _, h := range rule.ForbiddenHeaders {
		if req.Header.Get(h.Key) == h.Value {
			return false
		}
	}

	for _, ua := range rule.ForbiddenUserAgents {
		if ua.MatchString(req.UserAgent()) {
			return false
		}
	}

	if rule.MaxRequestLengthBytes > 0 && req.ContentLength > rule.MaxRequestLengthBytes {
		return false
	}

	bodyB, err := io.ReadAll(req.Body)
	if err != nil {
		return false
	}
	req.Body = io.NopCloser(bytes.NewBuffer(bodyB))

	for _, reqExpr := range rule.ForbiddenRequestRegexp {
		if reqExpr.Match(bodyB) {
			return false
		}
	}

	return true
}

func (f *Firewall) CheckResponse(endpoint string, resp *http.Response) (ok bool) {
	rule := f.endpointRules[endpoint]

	for _, h := range rule.RequiredHeaders {
		if len(resp.Header.Get(h)) == 0 {
			return false
		}
	}

	for _, h := range rule.ForbiddenHeaders {
		if resp.Header.Get(h.Key) == h.Value {
			return false
		}
	}

	if rule.MaxResponseLengthBytes > 0 && resp.ContentLength > rule.MaxResponseLengthBytes {
		return false
	}

	if slices.Contains(rule.ForbiddenResponseCodes, resp.StatusCode) {
		return false
	}

	bodyB, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyB))

	for _, reqExpr := range rule.ForbiddenResponseRegexp {
		if reqExpr.Match(bodyB) {
			return false
		}
	}

	return true
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config file %v", err)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading config file %v", err)
	}
	var conf Config
	if err := yaml.Unmarshal(b, &conf); err != nil {
		return nil, fmt.Errorf("unmarshaling config %v", err)
	}
	return &conf, nil
}

func CompileSliceRegexp(s []string) ([]*regexp.Regexp, error) {
	res := make([]*regexp.Regexp, 0, len(s))
	for _, raw := range s {
		expr, err := regexp.Compile(raw)
		if err != nil {
			return nil, err
		}
		res = append(res, expr)
	}
	return res, nil
}

type transport struct {
	http.RoundTripper
	Firewall *Firewall
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	fbd := &http.Response{
		StatusCode: http.StatusForbidden,
		Body:       io.NopCloser(bytes.NewBufferString("Forbidden")),
	}

	if ok := t.Firewall.CheckRequest(req.URL.Path, req); !ok {
		return fbd, nil
	}

	resp, err := t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if ok := t.Firewall.CheckResponse(req.URL.Path, resp); !ok {
		return fbd, nil
	}

	return resp, nil
}

func main() {
	srvAddr := flag.String("service-addr", "", "address of protected service")
	listenAddr := flag.String("addr", "", "address to run firewall on")
	confPath := flag.String("conf", "", "path to firewall config")
	flag.Parse()

	conf, err := LoadConfig(*confPath)
	if err != nil {
		log.Fatalf("loading config %v", err)
	}

	fw, err := New(conf)
	if err != nil {
		log.Fatalf("creating firewall %v", err)
	}

	srvURL, err := url.Parse(*srvAddr)
	if err != nil {
		log.Fatalf("parsing url %v", err)
	}

	rp := httputil.NewSingleHostReverseProxy(srvURL)
	rp.Transport = &transport{RoundTripper: http.DefaultTransport, Firewall: fw}

	s := http.Server{
		Addr:    *listenAddr,
		Handler: rp,
	}
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("running server %v", err)
	}
}
