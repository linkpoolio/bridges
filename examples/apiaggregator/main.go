package main

import (
	"errors"
	"fmt"
	"github.com/linkpoolio/bridges/bridge"
	"github.com/montanaflynn/stats"
	"github.com/oliveagle/jsonpath"
	"net/http"
	"strconv"
	"sync"
)

// Result represents the resulting data returned to Chainlink and
// merged in `data`
type Result struct {
	AggregationType string   `json:"aggregationType"`
	AggregateValue  string   `json:"aggregateValue"`
	FailedAPICount  int      `json:"failedApiCount"`
	APIErrors       []string `json:"apiErrors"`
}

// APIAggregator is a bridge that allows any public API that return numerical
// values to be aggregated by different types, currently supporting:
//  - Mean
//  - Median
//  - Mode
// To use the bridge:
//  - `api` []string List of APIs to query
//  - `paths` []string JSON paths to parse the returning responses
//  - `type` string Aggregation type to use
// For example:
// {
//    "api": [
//        "https://www.bitstamp.net/api/v2/ticker/btcusd/",
//        "https://api.pro.coinbase.com/products/btc-usd/ticker"
//    ],
//    "paths": ["$.last", "$.price"],
//    "type": "median"
// }
type APIAggregator struct{}

// Run is the bridge.Bridge Run implementation that returns the aggregated result
func (cc *APIAggregator) Run(h *bridge.Helper) (interface{}, error) {
	al := len(h.Data.Get("api").Array())
	pl := len(h.Data.Get("paths").Array())

	var wg sync.WaitGroup
	wg.Add(al)

	values := make(chan float64, al)
	errs := make(chan error, al)

	if (al == 0 && pl == 0) || al != pl {
		return h, errors.New("Invalid api and path array")
	}

	p := h.Data.Get("paths").Array()
	for i, a := range h.Data.Get("api").Array() {
		go performRequest(h, &wg, a.String(), p[i].String(), values, errs)
	}
	wg.Wait()
	close(values)
	close(errs)

	var r Result
	r.AggregationType = h.GetParam("type")
	if len(errs) > 0 {
		r.FailedAPICount = len(errs)
		var ehAh []string
		for err := range errs {
			ehAh = append(ehAh, err.Error())
		}
		r.APIErrors = ehAh
	}

	if aggValue, eh := aggregateValues(r.AggregationType, values); eh != nil {
		return nil, fmt.Errorf("Error aggregating value: %s", eh)
	} else {
		r.AggregateValue = aggValue
	}

	return r, nil
}

// Opts is the bridge.Bridge implementation
func (cc *APIAggregator) Opts() *bridge.Opts {
	return &bridge.Opts{
		Name:   "APIAggregator",
		Lambda: true,
	}
}

func main() {
	bridge.NewServer(&APIAggregator{}).Start(8080)
}

func performRequest(
	h *bridge.Helper,
	wg *sync.WaitGroup,
	api string,
	path string,
	values chan<- float64,
	errs chan<- error,
) {
	var obj interface{}
	defer wg.Done()

	if err := h.HTTPCall(http.MethodGet, api, &obj); err != nil {
		errs <- err
		return
	}
	val, err := jsonpath.JsonPathLookup(obj, path)
	if err != nil {
		errs <- err
		return
	}
	fv, err := strconv.ParseFloat(fmt.Sprint(val), 64)
	if err != nil {
		errs <- err
		return
	}

	values <- fv
}

func aggregateValues(aggType string, values chan float64) (string, error) {
	var av float64
	var eh error
	var valAh []float64

	for v := range values {
		valAh = append(valAh, v)
	}

	switch aggType {
	case "mode":
		var modeAh []float64
		modeAh, eh = stats.Mode(valAh)
		if len(modeAh) == 0 {
			av = valAh[0]
		} else {
			av = modeAh[0]
		}
		break
	case "median":
		av, eh = stats.Median(valAh)
		break
	default:
		av, eh = stats.Mean(valAh)
		break
	}

	return fmt.Sprint(av), eh
}
