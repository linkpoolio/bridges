package main

import (
	"github.com/linkpoolio/bridges"
	"net/http"
)

type GasStation struct{}

// Run implements Bridge Run for querying the Wolfram short answers API
func (gs *GasStation) Run(h *bridges.Helper) (interface{}, error) {
	obj := make(map[string]interface{})
	err := h.HTTPCall(
		http.MethodGet,
		"https://ethgasstation.info/json/ethgasAPI.json",
		&obj,
	)
	return obj["average"], err
}

// Opts is the bridge.Bridge implementation
func (gs *GasStation) Opts() *bridges.Opts {
	return &bridges.Opts{
		Name:   "GasStation",
		Lambda: true,
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	bridges.NewServer(&GasStation{}).Handler(w, r)
}

func main() {
	bridges.NewServer(&GasStation{}).Start(8080)
}
