package main

import (
	"errors"
	"github.com/jleeh/bridges/bridge"
	"net/http"
	"os"
	"strings"
)

// WolframAlpha is a bridge that queries the WolframAlpha short answers API
// https://products.wolframalpha.com/short-answers-api/documentation/
//
// This providers a natural way for a Chainlink end user to query data.
// For example:
//
// {
//    "query": "How far is Los Angeles from New York?",
//    "index": 1
// }
//
// Returning:
//
// {
//    "full": "about 2464 miles",
//    "result": "2464",
//    "unit": "miles"
// }
type WolframAlpha struct{}

// Run implements Bridge Run for querying the Wolfram short answers API
func (cc *WolframAlpha) Run(h *bridge.Helper) (interface{}, error) {
	b, err := h.HTTPCallRawWithOpts(
		http.MethodGet,
		"https://api.wolframalpha.com/v1/result",
		bridge.CallOpts{
			Auth: bridge.NewAuth(bridge.AuthParam, "appid", os.Getenv("APP_ID")),
			Param: map[string]interface{}{
				"i": h.GetParam("query"),
			},
		},
	)
	val := strings.Split(string(b), " ")
	i := h.GetIntParam("index")
	if i > int64(len(val)-1) {
		return nil, errors.New("Invalid index")
	}
	return map[string]string{"result": val[i], "full": string(b), "unit": val[len(val)-1]}, err
}

// Opts is the bridge.Bridge implementation
func (cc *WolframAlpha) Opts() *bridge.Opts {
	return &bridge.Opts{
		Name:   "WolframAlpha",
		Lambda: true,
	}
}

func main() {
	bridge.NewServer(&WolframAlpha{}).Start(8080)
}
