package bridges

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"gopkg.in/guregu/null.v3"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	AuthParam  = "param"
	AuthHeader = "header"
)

// Opts is the options for each bridge
type Opts struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Lambda bool   `json:"Lambda"`
}

// Result represents a Chainlink JobRun
type Result struct {
	JobRunID  string      `json:"jobRunId"`
	ID        string      `json:"id,omitempty"`
	TaskRunID string      `json:"taskRunId,omitempty"`
	Status    string      `json:"status"`
	Error     null.String `json:"error"`
	Pending   bool        `json:"pending"`
	Data      *JSON       `json:"data"`
}

// Based on https://github.com/smartcontractkit/chainlink/blob/master/core/store/models/common.go#L128
type JSON struct {
	gjson.Result
}

// ParseInterface attempts to coerce the input interface
// and parse it into a JSON object.
func ParseInterface(obj interface{}) (*JSON, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}

// ParseInterface parse a bytes slice into a JSON object.
func Parse(b []byte) (*JSON, error) {
	var j JSON
	str := string(b)
	if len(str) == 0 {
		str = `{}`
	}
	err := json.Unmarshal([]byte(str), &j)
	return &j, err
}

// UnmarshalJSON parses the JSON bytes and stores in the *JSON pointer.
func (j *JSON) UnmarshalJSON(b []byte) error {
	str := string(b)
	if !gjson.Valid(str) {
		return fmt.Errorf("invalid JSON: %v", str)
	}
	*j = JSON{gjson.Parse(str)}
	return nil
}

// MarshalJSON returns the JSON data if it already exists, returns
// an empty JSON object as bytes if not.
func (j *JSON) MarshalJSON() ([]byte, error) {
	if j.Exists() {
		return []byte(j.String()), nil
	}
	return []byte("{}"), nil
}

// SetCompleted marks a result as errored
func (r *Result) SetErrored(err error) {
	r.Status = "errored"
	r.Error = null.StringFrom(err.Error())
}

// SetCompleted marks a result as completed
func (r *Result) SetCompleted() {
	r.Status = "completed"
}

// SetJobRunID sets the request's ID to the result's Job Run ID.
// If "jobRunId" is supplied in the request, use that for the response.
func (r *Result) SetJobRunID() {
	if len(r.JobRunID) == 0 {
		r.JobRunID = r.ID
	}
}

// Bridge is the interface that can be implemented for custom Chainlink bridges
type Bridge interface {
	Opts() *Opts
	Run(h *Helper) (interface{}, error)
}

// Server holds pointers to the bridges indexed by their paths
// and the bridge to be mounted in Lambda.
type Server struct {
	pathMap   map[string]Bridge
	ldaBridge Bridge
}

// NewServer returns a new Server with the bridges
// in a map indexed by its path.
// Once returned, the server can be started to listen
// for any new requests.
//
// If a bridge is passed in that has a duplicate path
// then the last one with that path will be mounted.
//
// Any bridge with an empty path gets assigned "/" to avoid
// panics on start.
func NewServer(bridges ...Bridge) *Server {
	pm := make(map[string]Bridge)
	var lda Bridge
	for _, b := range bridges {
		var p string
		c := b.Opts()
		if len(c.Path) == 0 {
			p = "/"
		} else {
			p = c.Path
		}
		pm[p] = b
		if c.Lambda && lda == nil {
			lda = b
		}
	}
	return &Server{
		pathMap:   pm,
		ldaBridge: lda,
	}
}

// Start the bridge server. Routing on how the server is started is determined which
// platform is specified by the end user. Currently supporting:
//  - Inbuilt http (default)
//  - AWS Lambda (env LAMBDA=1)
//
// Port only has to be passed in if the inbuilt HTTP server is being used.
//
// If the inbuilt http server is being used, bridges can specify many external adaptors
// as long if exclusive paths are given.
//
// If multiple adaptors are included with Lambda/gcp enabled, then the first bridge that
// has it enabled will be given as the Handler.
func (s *Server) Start(port int) {
	if len(os.Getenv("LAMBDA")) > 0 {
		lambda.Start(s.Lambda)
	} else {
		logrus.WithField("port", port).Info("Starting the bridge server")
		logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), s.Mux()))
	}
}

// HTTPMux returns the http.Handler for the go http multiplexer, registering all the bridges paths
// with the handler
func (s *Server) Mux() http.Handler {
	mux := http.NewServeMux()
	for p, b := range s.pathMap {
		logrus.WithField("path", p).WithField("bridge", b.Opts().Name).Info("Registering bridge")
		mux.HandleFunc(p, s.Handler)
	}
	return mux
}

// Hander is of http.Handler type, receiving any inbound requests from the HTTP server
// when the bridge is ran local
func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {
	var rt Result
	start := time.Now()
	cc := make(chan int, 1)

	defer func() {
		code := <-cc
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		if err := json.NewEncoder(w).Encode(&rt); err != nil {
			logrus.Errorf("Failed to encode response: %v", err)
		}
		s.logRequest(r, code, start)
	}()

	if r.Method != http.MethodPost {
		cc <- http.StatusBadRequest
		rt.SetErrored(errors.New("Invalid request"))
		return
	}
	if b, err := ioutil.ReadAll(r.Body); err != nil {
		cc <- http.StatusInternalServerError
		rt.SetErrored(err)
		return
	} else if err = json.Unmarshal(b, &rt); err != nil {
		cc <- http.StatusBadRequest
		rt.SetErrored(err)
		return
	}

	rt.SetJobRunID()

	if b, ok := s.pathMap[s.path(r)]; !ok {
		cc <- http.StatusBadRequest
		rt.SetErrored(errors.New("Invalid path"))
	} else if obj, err := b.Run(NewHelper(rt.Data)); err != nil {
		cc <- http.StatusInternalServerError
		rt.SetErrored(err)
	} else if data, err := ParseInterface(obj); err != nil {
		cc <- http.StatusInternalServerError
		rt.SetErrored(err)
	} else {
		rt.Data = data
		rt.SetCompleted()
		cc <- http.StatusOK
	}
}

func (s *Server) Lambda(r *Result) (interface{}, error) {
	r.SetJobRunID()

	if obj, err := s.ldaBridge.Run(NewHelper(r.Data)); err != nil {
		r.SetErrored(err)
	} else if data, err := ParseInterface(obj); err != nil {
		r.SetErrored(err)
	} else {
		r.SetCompleted()
		r.Data = data
	}
	return r, nil
}

func (s *Server) logRequest(r *http.Request, code int, start time.Time) {
	end := time.Now()
	logrus.WithFields(logrus.Fields{
		"method":   r.Method,
		"code":     code,
		"path":     r.URL.Path,
		"clientIP": r.RemoteAddr,
		"servedAt": end.Format("2006/01/02 - 15:04:05"),
		"latency":  fmt.Sprintf("%v", end.Sub(start)),
	}).Info("Bridge request")
}

// Transformative logic to prepare the path, as if it's empty, it needs
// setting to the root "/" path
func (s *Server) path(r *http.Request) string {
	path := r.URL.Path
	if len(path) == 0 {
		path = "/"
	}
	return path
}

// Helper is given to the receiving bridge to use on run, giving the
// bridge the visibility to the input parameters from the node request
// and having simple functions for making http calls.
type Helper struct {
	Data *JSON

	httpClient http.Client
}

func NewHelper(data *JSON) *Helper {
	return &Helper{Data: data, httpClient: http.Client{}}
}

// GetIntParam gets the string value of a key in the `data` JSON object that is
// given on request by the Chainlink node
func (h *Helper) GetParam(key string) string {
	return h.Data.Get(key).String()
}

// GetIntParam gets the int64 value of a key in the `data` JSON object that is
// given on request by the Chainlink node
func (h *Helper) GetIntParam(key string) int64 {
	return h.Data.Get(key).Int()
}

// CallOpts are the options given into a http call method
type CallOpts struct {
	Auth             Auth                   `json:"-"`
	Query            map[string]interface{} `json:"query"`
	QueryPassthrough bool                   `json:"queryPassthrough"`
	Body             string                 `json:"body"`
	ExpectedCode     int                    `json:"expectedCode"`
}

// HTTPCall performs a basic http call with no options
func (h *Helper) HTTPCall(method, url string, obj interface{}) error {
	return h.HTTPCallWithContext(context.Background(), method, url, obj)
}

func (h *Helper) HTTPCallWithContext(ctx context.Context, method, url string, obj interface{}) error {
	return h.HTTPCallWithOptsWithContext(ctx, method, url, obj, CallOpts{})
}

// HTTPCallWithOpts mirrors HTTPCallRawWithOpts bar the returning byte body is unmarshalled into
// a given object pointer
func (h *Helper) HTTPCallWithOpts(method, url string, obj interface{}, opts CallOpts) error {
	return h.HTTPCallWithOptsWithContext(context.Background(), method, url, obj, opts)
}

func (h *Helper) HTTPCallWithOptsWithContext(ctx context.Context, method, url string, obj interface{}, opts CallOpts) error {
	if b, err := h.HTTPCallRawWithOptsWithContext(ctx, method, url, opts); err != nil {
		return err
	} else if err := json.Unmarshal(b, obj); err != nil {
		return err
	}
	return nil
}

// HTTPCallRawWithOpts performs a HTTP call with any method and returns the raw byte body and any error
// Supported options:
//  - Authentication methods for the API (query param, headers)
// 	- Query parameters via `opts.Query`
//  - Passthrough through all json keys within the request `data` object via `opts.QueryPassthrough`
//  - Pass in a body to send with the request via `opts.Body`
//  - Send in post form kv via `opts.PostForm`
//  - Return an error if the returning http status code is different to `opts.ExpectedCode`
func (h *Helper) HTTPCallRawWithOpts(method, url string, opts CallOpts) ([]byte, error) {
	return h.HTTPCallRawWithOptsWithContext(context.Background(), method, url, opts)
}

func (h *Helper) HTTPCallRawWithOptsWithContext(ctx context.Context, method, url string, opts CallOpts) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader([]byte(opts.Body)))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	q := req.URL.Query()
	if opts.QueryPassthrough {
		for k, v := range h.Data.Map() {
			q.Add(k, fmt.Sprintf("%s", v))
		}
	} else {
		for k, v := range opts.Query {
			q.Add(k, fmt.Sprintf("%s", v))
		}
	}
	req.URL.RawQuery = q.Encode()
	if opts.Auth != nil {
		opts.Auth.Authenticate(req)
	}

	if resp, err := h.httpClient.Do(req); err != nil {
		return nil, err
	} else if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	} else if (opts.ExpectedCode != 0 && resp.StatusCode != opts.ExpectedCode) ||
		opts.ExpectedCode == 0 && resp.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected api status code: %d", resp.StatusCode)
	} else {
		return b, nil
	}
}

// Auth is the generic interface for how the client passes in their
// API key for authentication
type Auth interface {
	Authenticate(*http.Request)
}

// NewAuth returns a pointer of an Auth implementation based on the
// type that was passed in
func NewAuth(authType string, key string, value string) Auth {
	var a Auth
	switch authType {
	case AuthParam:
		a = &Param{Key: key, Value: value}
		break
	case AuthHeader:
		a = &Header{Key: key, Value: value}
	}
	return a
}

// Query is the Auth implementation that requires GET param set
type Param struct {
	Key   string
	Value string
}

// Authenticate takes the `apikey` in the GET param and then authenticates it
// with the KeyManager
func (p *Param) Authenticate(r *http.Request) {
	q := r.URL.Query()
	q.Add(p.Key, p.Value)
	r.URL.RawQuery = q.Encode()
}

// Header is the Auth implementation that requires a header to be set
type Header struct {
	Key   string
	Value string
}

// Authenticate takes the key and value given and sets it as a header
func (p *Header) Authenticate(r *http.Request) {
	r.Header.Add(p.Key, p.Value)
}
