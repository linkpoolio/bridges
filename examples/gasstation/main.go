package main

import (
	"github.com/linkpoolio/bridges/bridge"
	"net/http"
)

type GasStation struct{}

// Run implements Bridge Run for querying the Wolfram short answers API
func (gs *GasStation) Run(h *bridge.Helper) (interface{}, error) {
	obj := make(map[string]interface{})
	err := h.HTTPCall(
		http.MethodGet,
		"https://ethgasstation.info/json/ethgasAPI.json",
		&obj,
	)
	return obj["average"], err
}

// Opts is the bridge.Bridge implementation
func (gs *GasStation) Opts() *bridge.Opts {
	return &bridge.Opts{
		Name:   "GasStation",
		Lambda: true,
	}
}

func main() {
	bridge.NewServer(&GasStation{}).Start(8080)
}
